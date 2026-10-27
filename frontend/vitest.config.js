import { defineConfig, configDefaults } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'node',
    include: ['src/**/*.test.js'],
    exclude: [...configDefaults.exclude, 'e2e/**'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'json-summary', 'lcov'],
      reportsDirectory: './coverage',
      // Only pure business logic enters the count — not the DOM app controller.
      include: ['src/lib/**/*.js'],
      exclude: ['src/**/*.test.js', 'node_modules/**'],
      thresholds: {
        // Anchored to measured suite (~92% lines / ~88% branches). Gate sits
        // just below so a new untested rule fails the build without being unreachable.
        lines: 80,
        branches: 75,
        functions: 80,
        statements: 80,
      },
    },
  },
});
