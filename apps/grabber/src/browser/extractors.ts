import { Page } from 'playwright';
import { extractListingIdFromUrl } from '../util/ids.js';
import { logger } from '../logger.js';

export interface ExtractedCard {
  sourceListingId: string;
  url: string;
  title: string;
  priceText?: string;
  priceAmount?: number;
  currency?: string;
  locationText?: string;
  publisherDisplayName?: string;
  rawText?: string;
  position: number;
}

/**
 * Parses a price string and returns amount + currency.
 * Supports formats:
 * - ฿28,000
 * - 28,000 THB
 * - THB 28000
 * - $1,200
 */
export function parsePrice(text: string): { amount: number; currency: string } | null {
  if (!text) return null;

  const cleaned = text.trim();

  // Thai Baht with symbol: ฿28,000 or ฿ 28,000
  const thaiSymbol = cleaned.match(/฿\s*([\d,]+(?:\.\d+)?)/);
  if (thaiSymbol) {
    const amount = parseFloat(thaiSymbol[1].replace(/,/g, ''));
    if (!isNaN(amount)) return { amount, currency: 'THB' };
  }

  // Currency code after number: 28,000 THB
  const codeAfter = cleaned.match(/^([\d,]+(?:\.\d+)?)\s*([A-Z]{3})$/);
  if (codeAfter) {
    const amount = parseFloat(codeAfter[1].replace(/,/g, ''));
    if (!isNaN(amount)) return { amount, currency: codeAfter[2] };
  }

  // Currency code before number: THB 28000
  const codeBefore = cleaned.match(/^([A-Z]{3})\s*([\d,]+(?:\.\d+)?)$/);
  if (codeBefore) {
    const amount = parseFloat(codeBefore[2].replace(/,/g, ''));
    if (!isNaN(amount)) return { amount, currency: codeBefore[1] };
  }

  // USD or other symbol: $1,200
  const dollarSymbol = cleaned.match(/^\$\s*([\d,]+(?:\.\d+)?)/);
  if (dollarSymbol) {
    const amount = parseFloat(dollarSymbol[1].replace(/,/g, ''));
    if (!isNaN(amount)) return { amount, currency: 'USD' };
  }

  // Try bare number as fallback
  const bare = cleaned.match(/^([\d,]+(?:\.\d+)?)$/);
  if (bare) {
    const amount = parseFloat(bare[1].replace(/,/g, ''));
    if (!isNaN(amount)) return { amount, currency: '' };
  }

  return null;
}

/**
 * Extracts listing cards from the current Marketplace search page.
 * Uses conservative, generic extraction rather than relying on brittle class names.
 */
export async function extractListingCards(page: Page, maxCards: number): Promise<ExtractedCard[]> {
  const cards: ExtractedCard[] = [];

  try {
    // Find all anchor links that match the Facebook Marketplace item URL pattern
    const anchors = await page.$$('a[href*="/marketplace/item/"]');

    logger.debug({ count: anchors.length }, 'found marketplace item anchors');

    const seen = new Set<string>();

    for (let i = 0; i < Math.min(anchors.length, maxCards); i++) {
      try {
        const anchor = anchors[i];
        if (!anchor) continue;

        const href = await anchor.getAttribute('href');
        if (!href) continue;

        // Build full URL
        const url = href.startsWith('http') ? href : `https://www.facebook.com${href}`;

        // Extract listing ID from URL
        const sourceListingId = extractListingIdFromUrl(url);
        if (!sourceListingId) {
          logger.debug({ url }, 'skipping anchor: could not extract listing ID');
          continue;
        }

        // Skip duplicates
        if (seen.has(sourceListingId)) continue;
        seen.add(sourceListingId);

        // Get all visible text within this anchor and its parent container
        const rawText = await anchor.evaluate((el) => {
          // Walk up to find a reasonable container (max 3 levels)
          let container: Element = el;
          for (let level = 0; level < 3; level++) {
            if (container.parentElement) {
              container = container.parentElement;
            }
          }
          return container.textContent?.trim() ?? '';
        });

        // Parse title, price, location from raw text
        const lines = rawText
          .split(/\n|\r/)
          .map((l: string) => l.trim())
          .filter((l: string) => l.length > 0);

        // The first meaningful text that's not a price is usually the title
        let title = '';
        let priceText = '';
        let locationText = '';

        for (const line of lines) {
          if (!title && line.length > 3 && !line.match(/^฿|^\$|^\d/)) {
            title = line;
          } else if (!priceText && line.match(/฿|THB|USD|\$|\d+,\d{3}/)) {
            priceText = line;
          } else if (!locationText && title && priceText && line.length > 1) {
            locationText = line;
          }
        }

        // If we couldn't find a title, use the aria-label or first text chunk
        if (!title && lines.length > 0) {
          title = lines[0] ?? '';
        }

        if (!title) {
          logger.debug({ url, sourceListingId }, 'skipping card: no title found');
          continue;
        }

        // Parse price
        let priceAmount: number | undefined;
        let currency: string | undefined;
        if (priceText) {
          const parsed = parsePrice(priceText);
          if (parsed) {
            priceAmount = parsed.amount;
            currency = parsed.currency || undefined;
          }
        }

        cards.push({
          sourceListingId,
          url,
          title: title.slice(0, 500), // cap length
          priceText: priceText || undefined,
          priceAmount,
          currency,
          locationText: locationText || undefined,
          rawText: rawText.slice(0, 2000), // cap length
          position: cards.length + 1,
        });
      } catch (err) {
        logger.warn({ err, index: i }, 'error extracting card, skipping');
      }
    }
  } catch (err) {
    logger.error({ err }, 'error extracting listing cards');
  }

  return cards;
}
