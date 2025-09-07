const request = require('supertest');
const express = require('express');

// Mock the Category model
const mockSave = jest.fn();
const mockFindOne = jest.fn();
const mockFind = jest.fn();
const mockFindById = jest.fn();
const mockFindByIdAndUpdate = jest.fn();
const mockFindByIdAndDelete = jest.fn();

const MockCategory = jest.fn().mockImplementation((data) => ({
  ...data,
  save: mockSave
}));

MockCategory.findOne = mockFindOne;
MockCategory.find = mockFind;
MockCategory.findById = mockFindById;
MockCategory.findByIdAndUpdate = mockFindByIdAndUpdate;
MockCategory.findByIdAndDelete = mockFindByIdAndDelete;

// Mock the Category module
jest.mock('../src/models/Category', () => MockCategory);

// Create Express app for testing (bypassing the file system issue)
function createTestApp() {
  const app = express();
  app.use(express.json());
  app.use(express.urlencoded({ extended: true }));
  
  // Import and use routes inline to avoid file system issues
  const {
    createCategory,
    getCategories,
    getCategory,
    updateCategory,
    deleteCategory
  } = require('../src/controllers/categoryController');
  
  app.post('/api/categories', createCategory);
  app.get('/api/categories', getCategories);
  app.get('/api/categories/:id', getCategory);
  app.put('/api/categories/:id', updateCategory);
  app.delete('/api/categories/:id', deleteCategory);
  
  return app;
}

describe('Categories API', () => {
  let app;

  beforeEach(() => {
    jest.clearAllMocks();
    app = createTestApp();
  });

  describe('POST /categories', () => {
    it('should create a new category successfully', async () => {
      const savedCategory = {
        _id: 'category123',
        name: 'Food',
        limit: 500,
        description: 'Groceries and dining'
      };

      mockFindOne.mockResolvedValue(null);
      mockSave.mockResolvedValue(savedCategory);

      const response = await request(app)
        .post('/api/categories')
        .send({
          name: 'Food',
          limit: 500,
          description: 'Groceries and dining'
        });

      expect(response.status).toBe(201);
      expect(response.body.message).toBe('Category created successfully');
      expect(response.body.category).toBeDefined();
    });

    it('should return 409 if category already exists', async () => {
      const existingCategory = { name: 'Food' };
      mockFindOne.mockResolvedValue(existingCategory);

      const response = await request(app)
        .post('/api/categories')
        .send({
          name: 'Food',
          limit: 500
        });

      expect(response.status).toBe(409);
      expect(response.body.error).toBe('Category with this name already exists');
    });
  });

  describe('GET /categories', () => {
    it('should return all categories', async () => {
      const mockCategories = [
        { _id: '1', name: 'Food', limit: 500 },
        { _id: '2', name: 'Transport', limit: 300 }
      ];

      mockFind.mockReturnValue({
        sort: jest.fn().mockResolvedValue(mockCategories)
      });

      const response = await request(app).get('/api/categories');

      expect(response.status).toBe(200);
      expect(response.body.count).toBe(2);
      expect(response.body.categories).toHaveLength(2);
    });
  });

  describe('PUT /categories/:id', () => {
    it('should update a category successfully', async () => {
      const updatedCategory = {
        _id: 'category123',
        name: 'Updated Food',
        limit: 600
      };

      mockFindByIdAndUpdate.mockResolvedValue(updatedCategory);

      const response = await request(app)
        .put('/api/categories/category123')
        .send({
          name: 'Updated Food',
          limit: 600
        });

      expect(response.status).toBe(200);
      expect(response.body.message).toBe('Category updated successfully');
    });
  });

  describe('DELETE /categories/:id', () => {
    it('should delete a category successfully', async () => {
      const deletedCategory = { _id: 'category123', name: 'Food' };
      mockFindByIdAndDelete.mockResolvedValue(deletedCategory);

      const response = await request(app)
        .delete('/api/categories/category123');

      expect(response.status).toBe(200);
      expect(response.body.message).toBe('Category deleted successfully');
    });
  });
});
