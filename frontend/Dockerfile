# Stage 1: Build Stage
FROM node:18 AS builder

# Define the build argument
ARG NEXT_PUBLIC_BASE_API_URL

# Set the working directory
WORKDIR /app

# Copy package.json and package-lock.json to leverage caching
COPY package*.json ./

# Install dependencies
RUN npm ci

# Set the build-time environment variable
ENV NEXT_PUBLIC_BASE_API_URL=$NEXT_PUBLIC_BASE_API_URL

# Copy the rest of the application code
COPY . .

# Build the application
RUN npm run build

# Stage 2: Production Stage
FROM node:18-slim AS runner

# Set the working directory
WORKDIR /app

# Copy only the built files and necessary assets from the previous stage
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./

# Set environment variable to production
ENV NODE_ENV=production

# Expose port
EXPOSE 3000

# Start the application
CMD ["npm", "start"]
