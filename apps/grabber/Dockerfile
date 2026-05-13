FROM mcr.microsoft.com/playwright:v1.50.1-jammy

WORKDIR /app

# Copy package files
COPY apps/grabber/package*.json ./

# Install dependencies
RUN npm ci

# Copy source
COPY apps/grabber/ ./

# Build TypeScript
RUN npm run build

# Create directories for browser profile and snapshots
RUN mkdir -p /browser-profile /snapshots

CMD ["node", "dist/index.js"]
