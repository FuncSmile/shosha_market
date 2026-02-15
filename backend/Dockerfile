# Backend Dockerfile untuk ShoshaMart
FROM node:20-alpine AS base

# Install dependencies only when needed
FROM base AS deps

# Install OpenSSL, Python, and build tools for Prisma and native dependencies
RUN apk add --no-cache \
    libc6-compat \
    openssl \
    openssl-dev \
    python3 \
    make \
    g++

WORKDIR /app

# Copy only package.json first (let npm generate lock file)
COPY package.json ./

# Install with verbose output
RUN npm install --legacy-peer-deps --no-audit --no-fund --verbose || \
    (echo "========== NPM INSTALL FAILED ==========" && \
     echo "Checking npm and node versions:" && \
     node --version && npm --version && \
     exit 1)

# Rebuild the source code only when needed
FROM base AS builder

# Install OpenSSL for Prisma
RUN apk add --no-cache openssl openssl-dev

WORKDIR /app

# Copy node_modules from deps
COPY --from=deps /app/node_modules ./node_modules

# Copy all source files
COPY . .

# Verify critical files exist (separate checks for clarity)
RUN echo "========== VERIFYING BUILD CONTEXT ==========" && \
    ls -la && \
    echo "" && \
    echo "Checking required files..." && \
    if [ ! -f tsconfig.json ]; then echo "✗ ERROR: tsconfig.json missing"; exit 1; else echo "✓ tsconfig.json exists"; fi && \
    if [ ! -f nest-cli.json ]; then echo "✗ ERROR: nest-cli.json missing"; exit 1; else echo "✓ nest-cli.json exists"; fi && \
    if [ ! -d src ]; then echo "✗ ERROR: src directory missing"; exit 1; else echo "✓ src directory exists"; fi && \
    if [ ! -d prisma ]; then echo "✗ ERROR: prisma directory missing"; exit 1; else echo "✓ prisma directory exists"; fi && \
    echo "========== ALL FILES VERIFIED ✓ =========="

# Generate Prisma Client (Prisma 7: DATABASE_URL configured via prisma.config.ts)
# Using placeholder URL for TypeScript types during build
ENV DATABASE_URL="postgresql://placeholder:placeholder@localhost:5432/placeholder?schema=public"
RUN npx prisma generate || \
    (echo "========== PRISMA GENERATE FAILED ==========" && \
     echo "Checking Prisma schema:" && \
     cat prisma/schema.prisma && \
     echo "" && \
     echo "Checking Prisma config:" && \
     cat prisma.config.ts && \
     exit 1)

# Build application with verbose output
RUN npm run build || \
    (echo "========== BUILD FAILED ==========" && \
     echo "Checking TypeScript config:" && \
     cat tsconfig.json && \
     echo "Checking NestJS config:" && \
     cat nest-cli.json && \
     echo "Listing src directory:" && \
     ls -la src/ && \
     echo "Checking for syntax errors:" && \
     npx tsc --noEmit || true && \
     exit 1)

# Verify build output
RUN test -d dist && echo "✓ Build successful - dist directory created" || \
    (echo "✗ Build failed - no dist directory" && exit 1)

# Production image, copy all the files and run nest
FROM base AS runner

# Install OpenSSL for Prisma runtime
RUN apk add --no-cache openssl libstdc++

WORKDIR /app

ENV NODE_ENV=production

# Create non-root user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nestjs

# Copy built application
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/prisma ./prisma
COPY --from=builder /app/prisma.config.ts ./prisma.config.ts

# Change ownership
RUN chown -R nestjs:nodejs /app

USER nestjs

EXPOSE 3001

ENV PORT=3001

CMD ["node", "dist/src/main"]
