const express = require('express');
const router = express.Router();
const { getSummary, getCategoryReport } = require('../controllers/summaryController');
 
router.get('/summary/:month', getSummary);
router.get('/reports/category/:categoryId', getCategoryReport);
 
module.exports = router;
