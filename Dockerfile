# Build stage
FROM node:18-alpine AS builder
 
WORKDIR /app
 
# Copy package files
COPY package*.json ./
COPY package-lock.json* ./
 
 
# Copy source code
COPY . .
 
# Create non-root user for security
RUN addgroup -g 1001 -S nodejs
RUN adduser -S nextjs -u 1001
 
# Change ownership of the app directory
RUN chown -R nextjs:nodejs /app
USER nextjs
 
# Expose application port
EXPOSE 3000
 
# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node healthcheck.js || exit 1
 
# Start the application
CMD ["npm", "start"]
