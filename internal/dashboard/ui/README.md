# GoVisual Dashboard (Preact)

Modern, fast dashboard built with Preact and shadcn-ui components.

## Development

### Prerequisites

- Node.js 18+
- npm or yarn

### Setup

```bash
# Install dependencies
npm install

# Build for production
npm run build

# Watch mode for development
npm run dev
```

## Architecture

- **Preact**: Lightweight React alternative (3KB)
- **shadcn-ui**: Modern, accessible UI components
- **Tailwind CSS**: Utility-first CSS framework
- **esbuild**: Fast JavaScript bundler
- **TypeScript**: Type safety

## Project Structure

```
src/
├── components/
│   ├── ui/              # shadcn-ui components
│   ├── RequestTable.tsx  # Request list component
│   ├── RequestDetails.tsx # Request details view
│   └── PerformanceProfiler.tsx # Performance profiling UI
├── lib/
│   ├── api.ts           # API client
│   └── utils.ts         # Utility functions
├── App.tsx              # Main application component
└── index.tsx            # Entry point
```

## Features

- **Real-time Updates**: Live request monitoring via SSE
- **Performance Profiling**: CPU, memory, and goroutine tracking
- **Flame Graphs**: Interactive D3.js visualization
- **Bottleneck Detection**: Automatic performance issue identification
- **Clean UI**: Modern, minimal design with no gradients
- **Fast**: Built with performance in mind

## Building for Production

The build process:

1. Compiles TypeScript to JavaScript
2. Bundles all dependencies with esbuild
3. Processes CSS with Tailwind
4. Outputs to `../static/` directory

The Go backend embeds these static files for distribution.
