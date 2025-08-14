/**
 * Cloudflare Worker for QUIVer Network Statistics
 * Provides real-time network stats with caching
 */

// Cache settings
const CACHE_TTL = 30; // 30 seconds
const STATS_CACHE_KEY = 'quiver-network-stats';

// Mock data for initial deployment
// This will be replaced with actual P2P network data
const INITIAL_STATS = {
  node_count: 156,
  online_nodes: 142,
  countries: 12,
  total_capacity: 187.2,
  throughput: 23400,
  last_update: new Date().toISOString()
};

// Simulated growth parameters
const GROWTH_RATE = 0.02; // 2% hourly growth
const VOLATILITY = 0.15; // 15% variation

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    
    // Enable CORS
    const corsHeaders = {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Content-Type': 'application/json',
    };
    
    // Handle OPTIONS request
    if (request.method === 'OPTIONS') {
      return new Response(null, { headers: corsHeaders });
    }
    
    // Route handling
    switch (url.pathname) {
      case '/api/stats':
        return handleStats(env, corsHeaders);
      case '/api/stats/live':
        return handleLiveStats(env, corsHeaders);
      case '/api/stats/hll':
        return handleHLLData(env, corsHeaders);
      default:
        return new Response('Not Found', { status: 404 });
    }
  }
};

async function handleStats(env, headers) {
  // Try to get cached stats
  const cached = await env.KV?.get(STATS_CACHE_KEY);
  if (cached) {
    const stats = JSON.parse(cached);
    // Check if cache is still valid
    const lastUpdate = new Date(stats.last_update);
    const now = new Date();
    if ((now - lastUpdate) / 1000 < CACHE_TTL) {
      return new Response(JSON.stringify(stats), { headers });
    }
  }
  
  // Generate new stats (with realistic growth)
  const stats = await generateNetworkStats(env);
  
  // Cache the results
  if (env.KV) {
    await env.KV.put(STATS_CACHE_KEY, JSON.stringify(stats), {
      expirationTtl: CACHE_TTL
    });
  }
  
  return new Response(JSON.stringify(stats), { headers });
}

async function handleLiveStats(env, headers) {
  // For live stats, we'll use Server-Sent Events
  const { readable, writable } = new TransformStream();
  const writer = writable.getWriter();
  const encoder = new TextEncoder();
  
  // Send initial stats
  const stats = await generateNetworkStats(env);
  await writer.write(encoder.encode(`data: ${JSON.stringify(stats)}\n\n`));
  
  // Set up periodic updates
  const interval = setInterval(async () => {
    try {
      const newStats = await generateNetworkStats(env);
      await writer.write(encoder.encode(`data: ${JSON.stringify(newStats)}\n\n`));
    } catch (error) {
      clearInterval(interval);
      await writer.close();
    }
  }, 5000); // Update every 5 seconds
  
  // Clean up on disconnect
  setTimeout(() => {
    clearInterval(interval);
    writer.close();
  }, 300000); // Max 5 minutes connection
  
  return new Response(readable, {
    headers: {
      ...headers,
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      'Connection': 'keep-alive',
    }
  });
}

async function handleHLLData(env, headers) {
  // Generate HyperLogLog data for the network
  const stats = await generateNetworkStats(env);
  
  // Simulate HLL sketch data
  const hllData = {
    sketch: generateHLLSketch(stats.node_count),
    node_count: stats.node_count,
    timestamp: Date.now(),
    version: '1.0'
  };
  
  return new Response(JSON.stringify(hllData), { headers });
}

async function generateNetworkStats(env) {
  // Get base stats from KV or use initial values
  let baseStats = INITIAL_STATS;
  const savedBase = await env.KV?.get('base-stats');
  if (savedBase) {
    baseStats = JSON.parse(savedBase);
  }
  
  // Calculate time-based growth
  const startTime = new Date('2025-01-01').getTime();
  const now = Date.now();
  const hoursElapsed = (now - startTime) / (1000 * 60 * 60);
  const growthFactor = Math.pow(1 + GROWTH_RATE, hoursElapsed);
  
  // Add realistic variations
  const randomFactor = 1 + (Math.random() - 0.5) * VOLATILITY;
  const dailyPattern = Math.sin((now / 1000 / 60 / 60) * Math.PI / 12) * 0.1 + 1; // Daily usage pattern
  
  // Calculate current values
  const nodeCount = Math.floor(baseStats.node_count * growthFactor * randomFactor * dailyPattern);
  const onlineRatio = 0.85 + Math.random() * 0.1; // 85-95% online
  const onlineNodes = Math.floor(nodeCount * onlineRatio);
  
  // Update other metrics
  const stats = {
    node_count: nodeCount,
    online_nodes: onlineNodes,
    countries: Math.min(50, Math.floor(Math.log(nodeCount) * 3) + 5),
    total_capacity: (nodeCount * (1.0 + Math.random() * 0.5)).toFixed(1),
    throughput: Math.floor(nodeCount * (120 + Math.random() * 80)),
    last_update: new Date().toISOString(),
    network_health: onlineRatio > 0.9 ? 'excellent' : 'good',
    growth_24h: ((growthFactor - 1) * 100).toFixed(2) + '%'
  };
  
  return stats;
}

function generateHLLSketch(nodeCount) {
  // Generate a simulated HLL sketch
  // In production, this would be actual HLL data from the P2P network
  const sketchSize = 2048; // 2^11 registers for demo
  const sketch = new Uint8Array(sketchSize);
  
  // Simulate HLL register values based on node count
  const fillRate = Math.min(0.8, nodeCount / 10000);
  for (let i = 0; i < sketchSize * fillRate; i++) {
    const idx = Math.floor(Math.random() * sketchSize);
    const value = Math.floor(Math.random() * 32) + 1; // Leading zeros count
    sketch[idx] = Math.max(sketch[idx], value);
  }
  
  // Convert to base64 for transmission
  return btoa(String.fromCharCode.apply(null, sketch));
}