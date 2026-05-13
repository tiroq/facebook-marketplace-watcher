import { describe, it, expect } from '@jest/globals';
import { extractListingIdFromUrl } from './ids.js';

describe('extractListingIdFromUrl', () => {
  it('extracts ID from standard marketplace item URL', () => {
    const url = 'https://www.facebook.com/marketplace/item/123456789/';
    expect(extractListingIdFromUrl(url)).toBe('123456789');
  });

  it('extracts ID from URL without trailing slash', () => {
    const url = 'https://www.facebook.com/marketplace/item/987654321';
    expect(extractListingIdFromUrl(url)).toBe('987654321');
  });

  it('extracts ID from URL with query params', () => {
    const url = 'https://www.facebook.com/marketplace/item/111222333/?ref=search&referral_code=null';
    expect(extractListingIdFromUrl(url)).toBe('111222333');
  });

  it('returns null for non-marketplace URLs', () => {
    const url = 'https://www.facebook.com/groups/some-group/';
    expect(extractListingIdFromUrl(url)).toBeNull();
  });

  it('returns null for empty string', () => {
    expect(extractListingIdFromUrl('')).toBeNull();
  });
});
