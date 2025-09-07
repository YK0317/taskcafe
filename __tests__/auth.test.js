const request = require('supertest'); 

const app = require('../src/app'); 

const User = require('../src/models/User'); 

const jwt = require('jsonwebtoken'); 

// Mock the User model 

jest.mock('../src/models/User'); 

describe('Authentication API', () => { 

  beforeEach(() => { 

    jest.clearAllMocks(); 

  }); 

  describe('POST /login', () => { 

    it('should return 400 if email is invalid', async () => { 

      const response = await request(app) 

        .post('/api/auth/login') 

        .send({ email: 'invalid-email', password: 'password123' }); 

      expect(response.status).toBe(400); 

      expect(response.body.errors).toContain('Please provide a valid email address'); 

    }); 

    it('should return 401 if credentials are invalid', async () => { 

      User.findOne.mockResolvedValue(null); 

      const response = await request(app) 

        .post('/api/auth/login') 

        .send({ email: 'test@example.com', password: 'wrongpassword' }); 

      expect(response.status).toBe(401); 

      expect(response.body.error).toBe('Invalid email or password'); 

    }); 

    it('should return JWT token on successful login', async () => { 

      const mockUser = { 

        _id: 'user123', 

        email: 'test@example.com', 

        validatePassword: jest.fn().mockResolvedValue(true) 

      }; 

      User.findOne.mockResolvedValue(mockUser); 

      const response = await request(app) 

        .post('/api/auth/login') 

        .send({ email: 'test@example.com', password: 'password123' }); 

      expect(response.status).toBe(200); 

      expect(response.body).toHaveProperty('token'); 

      expect(response.body).toHaveProperty('expiresIn', '1h'); 

    }); 

  }); 

  describe('POST /refresh-token', () => { 

    it('should return 400 if token is not provided', async () => { 

      const response = await request(app) 

        .post('/api/auth/refresh-token') 

        .send({}); 

      expect(response.status).toBe(400); 

      expect(response.body.error).toBe('Token is required'); 

    }); 

    it('should return new token when valid token is provided', async () => { 

      const validToken = jwt.sign( 

        { userId: 'user123', email: 'test@example.com' }, 

        'your_default_jwt_secret_here', 

        { expiresIn: '1h' } 

      ); 

      const response = await request(app) 

        .post('/api/auth/refresh-token') 

        .send({ token: validToken }); 

      expect(response.status).toBe(200); 

      expect(response.body).toHaveProperty('token'); 

      expect(response.body).toHaveProperty('expiresIn', '1h'); 
    }); 
  }); 
}); 