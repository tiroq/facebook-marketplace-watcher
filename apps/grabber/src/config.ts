export interface GrabberConfig {
  natsUrl: string;
  browserProfileDir: string;
  headless: boolean;
  searchBaseUrl: string;
  maxCardsPerRun: number;
  maxScrollsPerRun: number;
  snapshotDir: string;
}

function getEnv(key: string, defaultValue: string): string {
  return process.env[key] ?? defaultValue;
}

function getEnvInt(key: string, defaultValue: number): number {
  const v = process.env[key];
  if (v) {
    const n = parseInt(v, 10);
    if (!isNaN(n)) return n;
  }
  return defaultValue;
}

function getEnvBool(key: string, defaultValue: boolean): boolean {
  const v = process.env[key];
  if (v === 'true' || v === '1') return true;
  if (v === 'false' || v === '0') return false;
  return defaultValue;
}

export function loadConfig(): GrabberConfig {
  return {
    natsUrl: getEnv('NATS_URL', 'nats://localhost:4222'),
    browserProfileDir: getEnv('FB_BROWSER_PROFILE_DIR', './data/browser-profile'),
    headless: getEnvBool('FB_HEADLESS', false),
    searchBaseUrl: getEnv(
      'FB_SEARCH_BASE_URL',
      'https://www.facebook.com/marketplace/search/?query='
    ),
    maxCardsPerRun: getEnvInt('MAX_CARDS_PER_RUN', 50),
    maxScrollsPerRun: getEnvInt('MAX_SCROLLS_PER_RUN', 5),
    snapshotDir: getEnv('SNAPSHOT_DIR', '/snapshots'),
  };
}
