const request = require('supertest');
const app = require('../src/app');
const Category = require('../src/models/Category');
const Expense = require('../src/models/Expense');
 
jest.mock('../src/models/Category');
jest.mock('../src/models/Expense');
 
describe('Reports API', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });
 
  describe('GET /summary/:month', () => {
    it('should generate monthly summary report', async () => {
      const mockCategories = [
        { _id: 'cat1', name: 'Food', limit: 500, isActive: true },
        { _id: 'cat2', name: 'Transport', limit: 300, isActive: true }
      ];
 
      const mockExpenses = [
        { _id: 'exp1', categoryId: 'cat1', amount: 150, date: new Date('2024-01-15') },
        { _id: 'exp2', categoryId: 'cat1', amount: 200, date: new Date('2024-01-20') },
        { _id: 'exp3', categoryId: 'cat2', amount: 100, date: new Date('2024-01-10') }
      ];
 
      Category.find.mockResolvedValue(mockCategories);
      Expense.find.mockResolvedValue(mockExpenses);
 
      const response = await request(app).get('/api/summary/2024-01');
 
      expect(response.status).toBe(200);
      expect(response.body.month).toBe('2024-01');
      expect(response.body.summary).toHaveLength(2);
      expect(response.body.totals.totalSpent).toBe(450);
    });
 
    it('should return 400 for invalid month format', async () => {
      const response = await request(app).get('/api/summary/invalid-date');
      expect(response.status).toBe(400);
    });
  });
});
