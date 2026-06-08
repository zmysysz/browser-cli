package browser

import (
	"github.com/playwright-community/playwright-go"
)

// stealthInitScript returns JavaScript that runs before every page load to
// override Playwright/ChromeDriver automation fingerprints that Google and
// other login providers check.
func stealthInitScript() string {
	return `// ==UserScript==
// @name         playwright-stealth
// @description  Override automation detection vectors
// ==/UserScript==
(() => {
  'use strict';

  // 1. Kill navigator.webdriver — the single loudest Playwright giveaway.
  //    Override the getter so it never returns true.
  const webdriverGetter = Object.getOwnPropertyDescriptor(
    Navigator.prototype, 'webdriver'
  );
  if (webdriverGetter) {
    Object.defineProperty(Navigator.prototype, 'webdriver', {
      ...webdriverGetter,
      get: () => undefined,
    });
  }

  // 2. Fake navigator.plugins — add at least one entry so it doesn't
  //    return an empty array (headless / controlled browsers often have 0).
  const origPlugins = navigator.plugins;
  if (origPlugins && origPlugins.length === 0) {
    // Minimal PDF viewer mock
    const pdfPlugin = new Proxy(
      { name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer',
        description: 'Portable Document Format' },
      { get: (target, prop) => (typeof prop === 'string' && !isNaN(Number(prop))
        ? undefined : Reflect.get(target, prop)) }
    );
    const fakePluginsArray = [pdfPlugin];
    fakePluginsArray.item = (i) => fakePluginsArray[i] || null;
    fakePluginsArray.namedItem = (n) => fakePluginsArray.find(p => p.name === n) || null;
    fakePluginsArray.refresh = () => {};
    Object.defineProperty(navigator, 'plugins', {
      get: () => fakePluginsArray,
      configurable: true,
    });
  }

  // 3. Fake navigator.mimeTypes — add PDF mime type.
  if (navigator.mimeTypes && navigator.mimeTypes.length === 0) {
    const pdfMime = { type: 'application/pdf', suffix: 'pdf',
      description: 'Portable Document Format', enabledPlugin: null };
    const fakeMimeArray = [pdfMime];
    fakeMimeArray.item = (i) => fakeMimeArray[i] || null;
    fakeMimeArray.namedItem = (n) => fakeMimeArray.find(m => m.type === n) || null;
    Object.defineProperty(navigator, 'mimeTypes', {
      get: () => fakeMimeArray,
      configurable: true,
    });
  }

  // 4. Chrome runtime object — Google login checks for window.chrome.
  if (!window.chrome) {
    window.chrome = {
      runtime: { connect: () => {}, sendMessage: () => {} },
      loadTimes: () => {},
      csi: () => {},
      app: { isInstalled: false, InstallState: {}, RunningState: {} },
    };
  }

  // 5. Override Permissions.query to grant "notifications" automatically
  //    (headless Chrome often denies it by default).
  if (navigator.permissions && navigator.permissions.query) {
    const origQuery = navigator.permissions.query.bind(navigator.permissions);
    navigator.permissions.query = (desc) => {
      if (desc && desc.name === 'notifications') {
        return Promise.resolve({ state: 'granted', onchange: null });
      }
      return origQuery(desc);
    };
  }

  // 6. navigator.languages — set a realistic list.
  Object.defineProperty(navigator, 'languages', {
    get: () => ['en-US', 'en', 'zh-CN', 'zh'],
    configurable: true,
  });

  // 7. navigator.hardwareConcurrency — avoid telling sites this is a
  //    limited VPS or CI runner.
  Object.defineProperty(navigator, 'hardwareConcurrency', {
    get: () => 8,
    configurable: true,
  });

  // 8. WebGL vendor/renderer — headless Chromium reports the Mesa llvmpipe
  //    driver which is a dead giveaway. Overriding through WebGL is complex;
  //    instead inject a fake canvas fingerprint that masks it.
  const origToDataURL = HTMLCanvasElement.prototype.toDataURL;
  // Leave canvas fingerprinting alone — overriding it breaks Google's
  // own captcha challenge in some cases.

  console.debug('[stealth] automation patches applied');
})();
`
}

// stealthContextOptions returns BrowserNewContextOptions with realistic
// user-agent, locale, viewport, and other settings that help avoid
// browser fingerprinting / bot detection.
//
// IMPORTANT: we deliberately do NOT override UserAgent. Chrome's default UA
// matches its actual build version, so it's the most realistic option. An
// outdated UA like "Chrome/131" on a Chrome 149 build is itself a fingerprint.
func stealthContextOptions() playwright.BrowserNewContextOptions {
	opts := playwright.BrowserNewContextOptions{
		Locale:    playwright.String("en-US"),
		ColorScheme: playwright.ColorSchemeLight,
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
		// Accept downloads automatically
		AcceptDownloads: playwright.Bool(true),
		// Bypass CSP so we can inject our scripts
		BypassCSP: playwright.Bool(true),
	}
	return opts
}

// stealthChromiumArgs returns additional Chromium command-line flags that
// suppress automation indicators.
func stealthChromiumArgs() []string {
	return []string{
		"--disable-blink-features=AutomationControlled",
		// The following are already set in server.go, but included for completeness:
		// "--no-sandbox", "--disable-extensions", "--disable-default-apps",
	}
}

// ApplyStealth configures a BrowserContext with anti-detection measures.
// It sets context options and adds the stealth init script.
func ApplyStealth(context playwright.BrowserContext) error {
	return context.AddInitScript(playwright.Script{
		Content: playwright.String(stealthInitScript()),
	})
}