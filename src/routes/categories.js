const express = require('express'); 

const router = express.Router(); 

const { 

  createCategory, 

  getCategories, 

  getCategory, 

  updateCategory, 

  deleteCategory 

} = require('../controllers/categoryController'); 

  

// Category routes 

router.post('/categories', createCategory); 

router.get('/categories', getCategories); 

router.get('/categories/:id', getCategory); 

router.put('/categories/:id', updateCategory); 

router.delete('/categories/:id', deleteCategory); 

  

module.exports = router; 