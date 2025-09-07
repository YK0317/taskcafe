// Test setup file
// This file runs before each test suite

// Mock mongoose to prevent database connections during testing
// jest.mock('mongoose', () => ({
//   Schema: jest.fn().mockImplementation(() => ({
//     pre: jest.fn(),
//     statics: {},
//     methods: {}
//   })),
//   model: jest.fn(),
//   connect: jest.fn(),
//   connection: {
//     close: jest.fn()
//   }
// }));

// Mock console methods to reduce noise during testing
global.console = {
  ...console,
  // Uncomment to ignore specific console methods during tests
  // log: jest.fn(),
  // debug: jest.fn(),
  // info: jest.fn(),
  // warn: jest.fn(),
  // error: jest.fn(),
};

// Set test timeout
jest.setTimeout(10000);

// Global test helpers can be added here
global.testHelpers = {
  // Add any global test utilities here
};
