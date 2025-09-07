const Joi = require('joi'); 


const validateRegister = (req, res, next) => { 

  const schema = Joi.object({ 

    name: Joi.string().min(2).max(50).required() 

      .messages({ 

        'string.empty': 'Name is required', 

        'string.min': 'Name must be at least 2 characters long', 

        'string.max': 'Name cannot exceed 50 characters' 

      }), 

    email: Joi.string().email().required() 

      .messages({ 

        'string.empty': 'Email is required', 

        'string.email': 'Please provide a valid email address' 

      }), 

    password: Joi.string().min(6).required() 

      .messages({ 

        'string.empty': 'Password is required', 

        'string.min': 'Password must be at least 6 characters long' 

      }) 

  }); 


  const { error } = schema.validate(req.body, { abortEarly: false }); 


  if (error) { 

    const errorMessages = error.details.map(detail => detail.message); 

    return res.status(400).json({ errors: errorMessages }); 

  } 

  next(); 

}; 


const validateLogin = (req, res, next) => { 

  const schema = Joi.object({ 

    email: Joi.string().email().required() 

      .messages({ 

        'string.empty': 'Email is required', 

        'string.email': 'Please provide a valid email address' 

      }), 

    password: Joi.string().required() 

      .messages({ 

        'string.empty': 'Password is required' 

      }) 

  }); 

  const { error } = schema.validate(req.body, { abortEarly: false }); 


  if (error) { 

    const errorMessages = error.details.map(detail => detail.message); 

    return res.status(400).json({ errors: errorMessages }); 

  } 

  next(); 

}; 

module.exports = { validateRegister, validateLogin }; 