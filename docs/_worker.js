export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const hostname = url.hostname;
    
    // Map subdomains to directories
    const subdomainMap = {
      'api.quiver.network': '/api',
      'docs.quiver.network': '/docs',
      'explorer.quiver.network': '/explorer',
      'dashboard.quiver.network': '/dashboard',
      'security.quiver.network': '/security',
      'quicpair.quiver.network': '/quicpair',
      'playground.quiver.network': '/playground',
      'status.quiver.network': '/status',
      'blog.quiver.network': '/blog',
      'community.quiver.network': '/community',
      'cdn.quiver.network': '/cdn',
      'quiver.network': '',
      'www.quiver.network': ''
    };
    
    // Get the subdomain directory
    const subdir = subdomainMap[hostname] || '';
    
    // Construct the new pathname
    let pathname = url.pathname;
    if (subdir && !pathname.startsWith(subdir)) {
      pathname = subdir + pathname;
    }
    
    // Add index.html for directory paths
    if (pathname.endsWith('/')) {
      pathname += 'index.html';
    } else if (!pathname.includes('.')) {
      pathname += '/index.html';
    }
    
    // Create new request with modified path
    const modifiedRequest = new Request(url.origin + pathname, request);
    
    // Fetch from Pages
    const response = await env.ASSETS.fetch(modifiedRequest);
    
    // Return 404 if not found
    if (response.status === 404) {
      // Try without .html extension
      const withoutHtml = pathname.replace('/index.html', '').replace('.html', '');
      if (withoutHtml !== pathname) {
        const retryRequest = new Request(url.origin + withoutHtml, request);
        const retryResponse = await env.ASSETS.fetch(retryRequest);
        if (retryResponse.status !== 404) {
          return retryResponse;
        }
      }
    }
    
    return response;
  }
};