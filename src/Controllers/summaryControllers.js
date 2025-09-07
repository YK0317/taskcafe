const Category = require('../models/Category');
const Expense = require('../models/Expense');
 
exports.getSummary = async (req, res) => {
  try {
    const { month } = req.params;
    const { year = new Date().getFullYear() } = req.query;
 
    // Validate month format (YYYY-MM)
    const monthRegex = /^\d{4}-\d{2}$/;
    if (!monthRegex.test(month)) {
      return res.status(400).json({ error: 'Invalid month format. Use YYYY-MM' });
    }
 
    // Get all active categories
    const categories = await Category.find({ isActive: true });
    
    // Get expenses for the specified month
    const startDate = new Date(`${month}-01`);
    const endDate = new Date(startDate);
    endDate.setMonth(endDate.getMonth() + 1);
    endDate.setDate(0); // Last day of the month
 
    const expenses = await Expense.find({
      date: {
        $gte: startDate,
        $lte: endDate
      }
    });
 
    // Calculate summary for each category
    const summary = categories.map(category => {
      const categoryExpenses = expenses.filter(expense => 
        expense.categoryId.toString() === category._id.toString()
      );
 
      const totalSpent = categoryExpenses.reduce((sum, expense) => sum + expense.amount, 0);
      const remaining = category.limit - totalSpent;
      const utilizationPercentage = category.limit > 0 ? (totalSpent / category.limit) * 100 : 0;
 
      return {
        categoryId: category._id,
        categoryName: category.name,
        limit: category.limit,
        totalSpent,
        remaining,
        utilizationPercentage: utilizationPercentage.toFixed(2),
        isOverBudget: totalSpent > category.limit,
        expenseCount: categoryExpenses.length
      };
    });
 
    // Calculate overall totals
    const totalLimit = summary.reduce((sum, item) => sum + item.limit, 0);
    const totalSpent = summary.reduce((sum, item) => sum + item.totalSpent, 0);
    const totalRemaining = summary.reduce((sum, item) => sum + item.remaining, 0);
    const overallUtilization = totalLimit > 0 ? (totalSpent / totalLimit) * 100 : 0;
 
    res.json({
      month,
      year,
      summary,
      totals: {
        totalLimit,
        totalSpent,
        totalRemaining,
        overallUtilization: overallUtilization.toFixed(2),
        isOverallOverBudget: totalSpent > totalLimit
      },
      generatedAt: new Date().toISOString()
    });
  } catch (error) {
    console.error('Get summary error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
};
 
exports.getCategoryReport = async (req, res) => {
  try {
    const { categoryId } = req.params;
    const { startDate, endDate } = req.query;
 
    const category = await Category.findById(categoryId);
    if (!category) {
      return res.status(404).json({ error: 'Category not found' });
    }
 
    let dateFilter = {};
    if (startDate && endDate) {
      dateFilter.date = {
        $gte: new Date(startDate),
        $lte: new Date(endDate)
      };
    }
 
    const expenses = await Expense.find({
      categoryId,
      ...dateFilter
    }).sort({ date: -1 });
 
    const totalSpent = expenses.reduce((sum, expense) => sum + expense.amount, 0);
    const averageExpense = expenses.length > 0 ? totalSpent / expenses.length : 0;
 
    res.json({
      category: {
        id: category._id,
        name: category.name,
        limit: category.limit
      },
      period: {
        startDate: startDate || 'Beginning',
        endDate: endDate || 'Now'
      },
      statistics: {
        totalExpenses: expenses.length,
        totalSpent,
        averageExpense: averageExpense.toFixed(2),
        remaining: category.limit - totalSpent,
        utilizationPercentage: category.limit > 0 ? (totalSpent / category.limit) * 100 : 0
      },
      expenses
    });
  } catch (error) {
    console.error('Get category report error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
};
