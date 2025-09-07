@echo off
echo Simple Taskcafe Build Script for Windows

echo Installing Go dependencies...
go mod download
go mod tidy

echo Installing frontend dependencies...
cd frontend
npm ci --legacy-peer-deps
cd ..

echo Building frontend...
cd frontend
npm run build
cd ..

echo Building backend...
go build -o build/taskcafe.exe ./cmd/taskcafe

echo Build completed! Run: build/taskcafe.exe
