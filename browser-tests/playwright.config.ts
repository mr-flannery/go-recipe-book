import { defineConfig, devices } from '@playwright/test';

/**
 * Browser test mode controls which browsers/viewports to test:
 * - full: all browsers (chromium, firefox, webkit) + mobile viewports
 * - medium: chromium only, desktop + mobile viewports
 * - minimal: chromium desktop only (default)
 */
const testMode = process.env.BROWSER_TEST_MODE || 'minimal';

const desktopChromium = { name: 'chromium', use: { ...devices['Desktop Chrome'] } };
const desktopFirefox = { name: 'firefox', use: { ...devices['Desktop Firefox'] } };
const desktopWebkit = { name: 'webkit', use: { ...devices['Desktop Safari'] } };
const mobileChrome = { name: 'Mobile Chrome', use: { ...devices['Pixel 5'] } };
const mobileSafari = { name: 'Mobile Safari', use: { ...devices['iPhone 12'] } };

function getProjects() {
  switch (testMode) {
    case 'full':
      return [desktopChromium, desktopFirefox, desktopWebkit, mobileChrome, mobileSafari];
    case 'medium':
      return [desktopChromium, mobileChrome, mobileSafari];
    case 'minimal':
    default:
      return [desktopChromium];
  }
}

export default defineConfig({
  globalSetup: require.resolve('./global-setup'),
  testDir: '.',
  testMatch: '**/[0-9][0-9]-*.spec.ts',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: "100%",
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
  },
  projects: getProjects(),
});
