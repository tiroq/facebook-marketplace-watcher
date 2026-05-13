import { describe, it, expect } from '@jest/globals';
import { parsePrice } from './extractors.js';

describe('parsePrice', () => {
  it('parses Thai Baht with symbol', () => {
    const result = parsePrice('฿28,000');
    expect(result).toEqual({ amount: 28000, currency: 'THB' });
  });

  it('parses THB after number', () => {
    const result = parsePrice('28,000 THB');
    expect(result).toEqual({ amount: 28000, currency: 'THB' });
  });

  it('parses THB before number', () => {
    const result = parsePrice('THB 28000');
    expect(result).toEqual({ amount: 28000, currency: 'THB' });
  });

  it('parses USD with dollar sign', () => {
    const result = parsePrice('$1,200');
    expect(result).toEqual({ amount: 1200, currency: 'USD' });
  });

  it('returns null for empty string', () => {
    expect(parsePrice('')).toBeNull();
  });

  it('returns null for non-price text', () => {
    expect(parsePrice('Bangkok Thailand')).toBeNull();
  });

  it('handles decimal prices', () => {
    const result = parsePrice('฿28,000.50');
    expect(result).toEqual({ amount: 28000.50, currency: 'THB' });
  });
});
