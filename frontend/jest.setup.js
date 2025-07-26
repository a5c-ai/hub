import '@testing-library/jest-dom';
import { toHaveNoViolations } from 'jest-axe';

expect.extend(toHaveNoViolations);

// Polyfill ResizeObserver for Headless UI components
global.ResizeObserver = class {
  constructor(callback) {}
  disconnect() {}
  observe() {}
  unobserve() {}
};

// Mock matchMedia for components relying on prefers-color-scheme
if (typeof window.matchMedia !== 'function') {
  window.matchMedia = () => ({
    matches: false,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  });
}
