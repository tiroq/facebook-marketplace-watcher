import { AckPolicy, NatsConnection, JetStreamClient, JsMsg, StringCodec } from 'nats';
import { loadConfig } from './config.js';
import { logger } from './logger.js';
import { connectNats, ensureStream, buildEnvelope } from './nats.js';
import { subjects } from './subjects.js';
import { EnvelopeSchema, SearchRequestedDataSchema, ListingCardObservedData } from './events/schemas.js';
import { openPersistentContext } from './browser/context.js';
import { runSearch } from './browser/searchRunner.js';
import { newId } from './util/ids.js';

const sc = StringCodec();

async function publishEvent(
  js: JetStreamClient,
  subject: string,
  eventType: string,
  correlationId: string,
  data: unknown
): Promise<void> {
  const envelope = buildEnvelope(newId(), eventType, 'grabber', correlationId, data);
  const payload = sc.encode(JSON.stringify(envelope));
  await js.publish(subject, payload);
}

async function handleSearchRequested(
  msg: JsMsg,
  js: JetStreamClient,
  config: ReturnType<typeof loadConfig>
): Promise<void> {
  const raw = sc.decode(msg.data);

  // Parse and validate the envelope
  let envelope;
  try {
    envelope = EnvelopeSchema.parse(JSON.parse(raw));
  } catch (err) {
    logger.error({ err }, 'invalid event envelope, ignoring');
    msg.ack();
    return;
  }

  // Parse and validate the search requested payload
  let searchData;
  try {
    searchData = SearchRequestedDataSchema.parse(envelope.data);
  } catch (err) {
    logger.error({ err }, 'invalid search_requested payload, ignoring');
    msg.ack();
    return;
  }

  const sessionId = searchData.search_session_id ?? newId();
  const correlationId = envelope.correlation_id;

  logger.info(
    {
      query: searchData.query,
      sessionId,
      maxCards: searchData.max_cards,
      maxScrolls: searchData.max_scrolls,
    },
    'handling search.requested'
  );

  // Publish search started
  try {
    await publishEvent(js, subjects.searchStarted, 'search.started', correlationId, {
      search_session_id: sessionId,
      query: searchData.query,
      started_at: new Date().toISOString(),
    });
  } catch (err) {
    logger.error({ err }, 'failed to publish search.started');
  }

  // Open browser context
  let context;
  try {
    context = await openPersistentContext(config.browserProfileDir, config.headless);
  } catch (err) {
    const errorMsg = err instanceof Error ? err.message : String(err);
    logger.error({ err: errorMsg }, 'failed to open browser context');

    await publishEvent(js, subjects.searchFailed, 'search.failed', correlationId, {
      search_session_id: sessionId,
      query: searchData.query,
      error_code: 'BROWSER_OPEN_FAILED',
      error_message: errorMsg,
      failed_at: new Date().toISOString(),
    }).catch(() => {});

    msg.ack();
    return;
  }

  try {
    const result = await runSearch(context, {
      query: searchData.query,
      searchBaseUrl: config.searchBaseUrl,
      maxCards: Math.min(searchData.max_cards, config.maxCardsPerRun),
      maxScrolls: Math.min(searchData.max_scrolls, config.maxScrollsPerRun),
    });

    if (result.intervention) {
      // Do not bypass - report the intervention and stop
      logger.warn({ intervention: result.intervention, sessionId }, 'search stopped due to intervention');

      await publishEvent(js, subjects.searchFailed, 'search.failed', correlationId, {
        search_session_id: sessionId,
        query: searchData.query,
        error_code: result.intervention,
        error_message: `Intervention detected: ${result.intervention}. Manual action required.`,
        failed_at: new Date().toISOString(),
      }).catch(() => {});

      msg.ack();
      return;
    }

    if (result.error) {
      const errorMsg = result.error.message;
      await publishEvent(js, subjects.searchFailed, 'search.failed', correlationId, {
        search_session_id: sessionId,
        query: searchData.query,
        error_code: 'SEARCH_ERROR',
        error_message: errorMsg,
        failed_at: new Date().toISOString(),
      }).catch(() => {});

      msg.ack();
      return;
    }

    // Publish one ListingCardObserved event per card
    let publishedCount = 0;
    for (const card of result.cards) {
      const observedData: ListingCardObservedData = {
        search_session_id: sessionId,
        search_query_id: searchData.search_query_id,
        source: searchData.source ?? 'facebook_marketplace',
        source_listing_id: card.sourceListingId,
        observed_at: new Date().toISOString(),
        position: card.position,
        url: card.url,
        title: card.title,
        price_text: card.priceText,
        price_amount: card.priceAmount,
        currency: card.currency,
        location_text: card.locationText,
        raw_text: card.rawText,
        raw_json: {},
      };

      try {
        await publishEvent(
          js,
          subjects.listingCardObserved,
          'listing.card.observed',
          correlationId,
          observedData
        );
        publishedCount++;
      } catch (err) {
        logger.error(
          { err, sourceListingId: card.sourceListingId },
          'failed to publish listing.card.observed'
        );
      }
    }

    // Publish search completed
    await publishEvent(js, subjects.searchCompleted, 'search.completed', correlationId, {
      search_session_id: sessionId,
      query: searchData.query,
      cards_observed: publishedCount,
      completed_at: new Date().toISOString(),
    }).catch((err) => logger.error({ err }, 'failed to publish search.completed'));

    logger.info({ sessionId, cardsPublished: publishedCount }, 'search completed');
  } finally {
    await context.close().catch(() => {});
  }

  msg.ack();
}

async function main(): Promise<void> {
  logger.info('starting grabber');

  const config = loadConfig();

  let nc: NatsConnection;
  let js: JetStreamClient;

  try {
    nc = await connectNats(config.natsUrl);
    js = await ensureStream(nc);
  } catch (err) {
    logger.error({ err }, 'failed to connect to NATS');
    process.exit(1);
  }

  // Subscribe to search.requested using JetStream
  const jsm = await nc.jetstreamManager();

  // Create or update a durable consumer for grabber
  try {
    await jsm.consumers.add('FB_EVENTS', {
      name: 'grabber-search-consumer',
      durable_name: 'grabber-search-consumer',
      filter_subject: subjects.searchRequested,
      ack_policy: AckPolicy.Explicit,
    });
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : String(err);
    if (!msg.includes('consumer name already in use') && !msg.includes('already exists')) {
      logger.error({ err }, 'failed to create NATS consumer');
      process.exit(1);
    }
    logger.info('NATS consumer already exists');
  }

  const consumer = await js.consumers.get('FB_EVENTS', 'grabber-search-consumer');

  logger.info({ subject: subjects.searchRequested }, 'subscribed to search.requested');

  // Process messages one at a time (grabber is sequential by design)
  const messages = await consumer.consume({ max_messages: 1 });

  // Handle graceful shutdown
  process.on('SIGTERM', async () => {
    logger.info('received SIGTERM, shutting down');
    messages.stop();
    await nc.drain();
    process.exit(0);
  });

  process.on('SIGINT', async () => {
    logger.info('received SIGINT, shutting down');
    messages.stop();
    await nc.drain();
    process.exit(0);
  });

  for await (const msg of messages) {
    try {
      await handleSearchRequested(msg, js, config);
    } catch (err) {
      logger.error({ err }, 'unhandled error processing search.requested');
      msg.nak();
    }
  }
}

main().catch((err) => {
  logger.error({ err }, 'fatal error');
  process.exit(1);
});
