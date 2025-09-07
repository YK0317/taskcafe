const express = require('express'); 

const router = express.Router(); 

const { login, refreshToken } = require('../controllers/authController'); 

const { validateRegister, validateLogin } = require('../middleware/validationMiddleware'); 

  

// Apply validation middleware to routes 

router.post('/register', validateRegister, (req, res) => { 

  // Your registration logic here 

}); 

  

router.post('/login', validateLogin, login); 

router.post('/refresh-token', refreshToken); 

  

module.exports = router; 