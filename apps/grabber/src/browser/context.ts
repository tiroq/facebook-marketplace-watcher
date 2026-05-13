import { BrowserContext, Page, chromium } from 'playwright';
import { logger } from '../logger.js';

/**
 * Opens a persistent browser context using the user's browser profile.
 * Does NOT automate login - the user must log in manually.
 */
export async function openPersistentContext(
  profileDir: string,
  headless: boolean
): Promise<BrowserContext> {
  logger.info({ profileDir, headless }, 'opening browser context');

  const context = await chromium.launchPersistentContext(profileDir, {
    headless,
    viewport: { width: 1280, height: 900 },
    locale: 'en-US',
    timezoneId: 'Asia/Bangkok',
    args: ['--disable-blink-features=AutomationControlled', '--no-sandbox'],
    ignoreDefaultArgs: ['--enable-automation'],
  });

  return context;
}

/**
 * Checks if the current page is showing a login, checkpoint, or captcha page.
 * Returns the intervention type or null if the page is normal.
 */
export async function detectIntervention(page: Page): Promise<string | null> {
  const url = page.url();

  // Login page
  if (url.includes('login') || url.includes('/login.php')) {
    return 'LOGIN_REQUIRED';
  }

  // Checkpoint / account verification
  if (url.includes('checkpoint') || url.includes('/checkpoint/')) {
    return 'CHECKPOINT_REQUIRED';
  }

  // Check for common intervention indicators in the page
  try {
    const title = await page.title();
    const lowerTitle = title.toLowerCase();

    if (lowerTitle.includes('log in') || lowerTitle.includes('sign in')) {
      return 'LOGIN_REQUIRED';
    }

    if (lowerTitle.includes('security check') || lowerTitle.includes('captcha')) {
      return 'CAPTCHA_OR_INTERVENTION_REQUIRED';
    }
  } catch {
    // Ignore title errors
  }

  // Check for CAPTCHA iframe or common captcha indicators
  try {
    const captchaFrame = page.frames().find(
      (f) => f.url().includes('captcha') || f.url().includes('recaptcha')
    );
    if (captchaFrame) {
      return 'CAPTCHA_OR_INTERVENTION_REQUIRED';
    }
  } catch {
    // Ignore frame errors
  }

  return null;
}
