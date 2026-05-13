import { v4 as uuidv4 } from 'uuid';

/**
 * Extracts the source_listing_id from a Facebook Marketplace URL.
 * Example: https://www.facebook.com/marketplace/item/123456789/ -> "123456789"
 */
export function extractListingIdFromUrl(url: string): string | null {
  // Match /marketplace/item/{id}/ or /marketplace/item/{id}?...
  const match = url.match(/\/marketplace\/item\/(\d+)/);
  if (match && match[1]) {
    return match[1];
  }

  // Fallback: try to match any numeric segment that looks like an FB item ID
  const fallback = url.match(/\/(\d{10,})/);
  if (fallback && fallback[1]) {
    return fallback[1];
  }

  return null;
}

/**
 * Generates a new UUID v4.
 */
export function newId(): string {
  return uuidv4();
}

/**
 * Creates a simple hash of a string (for dedup purposes only, not cryptographic).
 */
export function simpleHash(s: string): string {
  let hash = 0;
  for (let i = 0; i < s.length; i++) {
    const char = s.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash; // Convert to 32bit integer
  }
  return Math.abs(hash).toString(16);
}
