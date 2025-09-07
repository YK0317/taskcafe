module.exports = {
  testEnvironment: 'node',
  coverageDirectory: 'coverage',
  collectCoverageFrom: [
    'src/**/*.js',
    '!src/**/*.test.js',
    '!src/index.js'
  ],
  coverageReporters: [
    'text',
    'lcov',
    'html'
  ],
  testMatch: [
    '**/__tests__/**/*.test.js'
  ],
  // setupFilesAfterEnv: ['<rootDir>/__tests__/setup.js'],
  testPathIgnorePatterns: [
    '/node_modules/'
  ]
};
