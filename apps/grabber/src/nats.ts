import { connect, NatsConnection, JetStreamClient, RetentionPolicy, StorageType } from 'nats';
import { logger } from './logger.js';
import { STREAM_NAME } from './subjects.js';

export async function connectNats(url: string): Promise<NatsConnection> {
  const nc = await connect({ servers: url, name: 'grabber' });
  logger.info({ natsUrl: url }, 'connected to NATS');
  return nc;
}

export async function ensureStream(nc: NatsConnection): Promise<JetStreamClient> {
  const js = nc.jetstream();
  const jsm = await nc.jetstreamManager();

  try {
    await jsm.streams.add({
      name: STREAM_NAME,
      subjects: ['fb.>'],
      retention: RetentionPolicy.Limits,
      storage: StorageType.File,
    });
    logger.info({ stream: STREAM_NAME }, 'NATS stream created');
  } catch (err: unknown) {
    // Stream already exists is acceptable
    const msg = err instanceof Error ? err.message : String(err);
    if (msg.includes('stream name already in use') || msg.includes('already exists')) {
      logger.info({ stream: STREAM_NAME }, 'NATS stream already exists');
    } else {
      throw err;
    }
  }

  return js;
}

export function buildEnvelope(
  eventId: string,
  eventType: string,
  sourceService: string,
  correlationId: string,
  data: unknown
) {
  return {
    schema: 'fb_event',
    schema_version: 1,
    event_id: eventId,
    event_type: eventType,
    source_service: sourceService,
    occurred_at: new Date().toISOString(),
    correlation_id: correlationId,
    data,
  };
}
