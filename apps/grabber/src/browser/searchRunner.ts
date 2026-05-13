import { BrowserContext } from 'playwright';
import { logger } from '../logger.js';
import { extractListingCards, ExtractedCard } from './extractors.js';
import { detectIntervention } from './context.js';

export interface SearchRunnerOptions {
  query: string;
  searchBaseUrl: string;
  maxCards: number;
  maxScrolls: number;
}

export interface SearchRunResult {
  cards: ExtractedCard[];
  intervention: string | null;
  error?: Error;
}

/**
 * Runs a single search on Facebook Marketplace.
 * Navigates to the search URL, scrolls, and extracts listing cards.
 */
export async function runSearch(
  context: BrowserContext,
  options: SearchRunnerOptions
): Promise<SearchRunResult> {
  const page = await context.newPage();

  try {
    const searchUrl = `${options.searchBaseUrl}${encodeURIComponent(options.query)}`;
    logger.info(
      { searchUrl, maxCards: options.maxCards, maxScrolls: options.maxScrolls },
      'navigating to search'
    );

    await page.goto(searchUrl, { waitUntil: 'domcontentloaded', timeout: 30000 });

    // Check for intervention before doing anything else
    const intervention = await detectIntervention(page);
    if (intervention) {
      logger.warn({ intervention, url: page.url() }, 'intervention detected, stopping search');
      return { cards: [], intervention };
    }

    // Wait for page to load content
    await page.waitForTimeout(3000);

    // Scroll to load more results
    for (let scroll = 0; scroll < options.maxScrolls; scroll++) {
      await page.evaluate(() => window.scrollBy(0, window.innerHeight * 2));
      await page.waitForTimeout(1500);

      // Check for intervention after each scroll
      const interventionAfterScroll = await detectIntervention(page);
      if (interventionAfterScroll) {
        logger.warn({ intervention: interventionAfterScroll }, 'intervention detected during scroll');
        return { cards: [], intervention: interventionAfterScroll };
      }
    }

    // Extract listing cards
    const cards = await extractListingCards(page, options.maxCards);
    logger.info({ cardsFound: cards.length, query: options.query }, 'extraction complete');

    return { cards, intervention: null };
  } catch (err) {
    const error = err instanceof Error ? err : new Error(String(err));
    logger.error({ err: error.message }, 'error during search run');
    return { cards: [], intervention: null, error };
  } finally {
    await page.close().catch(() => {});
  }
}
