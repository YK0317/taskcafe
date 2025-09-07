const jwt = require('jsonwebtoken'); 

const User = require('../models/User'); // Ensure this model exists 

  

const JWT_SECRET = process.env.JWT_SECRET || 'your_default_jwt_secret_here'; 

const JWT_EXPIRY = process.env.JWT_EXPIRY || '1h'; 

  

exports.login = async (req, res) => { 

  try { 

    const { email, password } = req.body; 

     

    // 1. Find user by email 

    const user = await User.findOne({ email }); 

    if (!user) { 

      return res.status(401).json({ error: 'Invalid email or password' }); 

    } 

  

    // 2. Validate password (assuming you have a method like validatePassword in your User model) 

    const isValidPassword = await user.validatePassword(password); 

    if (!isValidPassword) { 

      return res.status(401).json({ error: 'Invalid email or password' }); 

    } 

  

    // 3. Generate JWT Token 

    const token = jwt.sign( 

      { userId: user._id, email: user.email }, 

      JWT_SECRET, 

      { expiresIn: JWT_EXPIRY } 

    ); 

  

    // 4. Respond with token 

    res.json({  

      message: 'Login successful',  

      token, 

      expiresIn: JWT_EXPIRY 

    }); 

  

  } catch (error) { 

    console.error('Login error:', error); 

    res.status(500).json({ error: 'Internal server error during login' }); 

  } 

}; 

  

exports.refreshToken = (req, res) => { 

  try { 

    const { token } = req.body; 

  

    if (!token) { 

      return res.status(400).json({ error: 'Token is required' }); 

    } 

  

    // Verify the existing token 

    jwt.verify(token, JWT_SECRET, (err, decoded) => { 

      if (err) { 

        return res.status(403).json({ error: 'Invalid or expired token' }); 

      } 

  

      // Issue a new token with the same payload 

      const newToken = jwt.sign( 

        { userId: decoded.userId, email: decoded.email }, 

        JWT_SECRET, 

        { expiresIn: JWT_EXPIRY } 

      ); 

  

      res.json({  

        message: 'Token refreshed successfully', 

        token: newToken, 

        expiresIn: JWT_EXPIRY 

      }); 

    }); 

  } catch (error) { 

    console.error('Refresh token error:', error); 

    res.status(500).json({ error: 'Internal server error during token refresh' }); 

  } 

}; 