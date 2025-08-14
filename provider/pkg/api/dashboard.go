package api

// DashboardHTML is the embedded dashboard HTML
const DashboardHTML = `
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>QUIVer Provider Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            background: linear-gradient(135deg, #1e3c72 0%, #2a5298 100%);
            color: white;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        
        .container {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            padding: 40px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
            max-width: 800px;
            width: 100%;
            margin: 20px;
        }
        
        h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            text-align: center;
        }
        
        .subtitle {
            text-align: center;
            opacity: 0.9;
            margin-bottom: 40px;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 40px;
        }
        
        .stat-card {
            background: rgba(255, 255, 255, 0.1);
            border-radius: 15px;
            padding: 25px;
            text-align: center;
            transition: transform 0.3s;
        }
        
        .stat-card:hover {
            transform: translateY(-5px);
        }
        
        .stat-value {
            font-size: 2.5em;
            font-weight: bold;
            margin-bottom: 5px;
        }
        
        .stat-label {
            opacity: 0.8;
            font-size: 0.9em;
        }
        
        .earnings {
            color: #4ade80;
        }
        
        .status {
            text-align: center;
            margin-bottom: 30px;
        }
        
        .status-badge {
            display: inline-block;
            padding: 10px 20px;
            border-radius: 25px;
            background: rgba(74, 222, 128, 0.2);
            color: #4ade80;
            font-weight: 500;
        }
        
        .offline {
            background: rgba(239, 68, 68, 0.2);
            color: #ef4444;
        }
        
        .controls {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin-top: 30px;
        }
        
        button {
            padding: 12px 30px;
            border: none;
            border-radius: 25px;
            font-size: 1em;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.3s;
            background: rgba(255, 255, 255, 0.2);
            color: white;
        }
        
        button:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: translateY(-2px);
        }
        
        .primary {
            background: #4ade80;
            color: #1e3c72;
        }
        
        .primary:hover {
            background: #22c55e;
        }
        
        .network-info {
            margin-top: 30px;
            padding: 20px;
            background: rgba(255, 255, 255, 0.05);
            border-radius: 15px;
        }
        
        .network-info h3 {
            margin-bottom: 15px;
        }
        
        .network-stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
        }
        
        .network-stat {
            text-align: center;
        }
        
        .network-value {
            font-size: 1.5em;
            font-weight: bold;
        }
        
        .network-label {
            font-size: 0.8em;
            opacity: 0.8;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>QUIVer Provider</h1>
        <p class="subtitle">P2P AI推論ネットワーク</p>
        
        <div class="status">
            <span id="status-badge" class="status-badge">接続中...</span>
        </div>
        
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-value earnings">¥<span id="earnings">0</span></div>
                <div class="stat-label">本日の収益</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="requests">0</div>
                <div class="stat-label">処理リクエスト</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="uptime">0h</div>
                <div class="stat-label">稼働時間</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="model">llama3.2</div>
                <div class="stat-label">提供モデル</div>
            </div>
        </div>
        
        <div class="network-info">
            <h3>ネットワーク状態</h3>
            <div class="network-stats">
                <div class="network-stat">
                    <div class="network-value" id="peers">0</div>
                    <div class="network-label">接続ピア</div>
                </div>
                <div class="network-stat">
                    <div class="network-value" id="total-nodes">0</div>
                    <div class="network-label">総ノード数</div>
                </div>
                <div class="network-stat">
                    <div class="network-value" id="cpu">0%</div>
                    <div class="network-label">CPU使用率</div>
                </div>
                <div class="network-stat">
                    <div class="network-value" id="memory">0%</div>
                    <div class="network-label">メモリ使用率</div>
                </div>
            </div>
        </div>
        
        <div class="controls">
            <button id="start-btn" class="primary">開始</button>
            <button id="stop-btn">停止</button>
            <button id="settings-btn">設定</button>
        </div>
    </div>
    
    <script>
        let isRunning = false;
        
        async function updateStats() {
            try {
                const response = await fetch('http://localhost:8083/stats');
                const data = await response.json();
                
                document.getElementById('earnings').textContent = Math.floor(data.earnings || 0);
                document.getElementById('requests').textContent = data.requests || 0;
                document.getElementById('uptime').textContent = formatUptime(data.uptime || '0s');
                
                const statusBadge = document.getElementById('status-badge');
                if (data.network_status === 'P2P Connected') {
                    statusBadge.textContent = 'P2P接続済み';
                    statusBadge.classList.remove('offline');
                } else if (data.network_status === 'Offline') {
                    statusBadge.textContent = 'オフライン';
                    statusBadge.classList.add('offline');
                } else {
                    statusBadge.textContent = '接続中...';
                    statusBadge.classList.remove('offline');
                }
                
                // Update network stats
                const netResponse = await fetch('http://localhost:8082/api/stats');
                const netData = await netResponse.json();
                
                document.getElementById('peers').textContent = netData.connected_peers || 0;
                document.getElementById('total-nodes').textContent = netData.total_nodes || 0;
                document.getElementById('cpu').textContent = Math.round(data.cpu_usage || 0) + '%';
                document.getElementById('memory').textContent = Math.round(data.memory_usage || 0) + '%';
            } catch (error) {
                console.error('Failed to fetch stats:', error);
            }
        }
        
        function formatUptime(uptime) {
            // Parse Go duration string
            const match = uptime.match(/(\d+)h(\d+)m/);
            if (match) {
                return match[1] + 'h ' + match[2] + 'm';
            }
            return uptime;
        }
        
        document.getElementById('start-btn').addEventListener('click', async () => {
            await fetch('http://localhost:8083/start', { method: 'POST' });
            isRunning = true;
            updateButtons();
        });
        
        document.getElementById('stop-btn').addEventListener('click', async () => {
            await fetch('http://localhost:8083/stop', { method: 'POST' });
            isRunning = false;
            updateButtons();
        });
        
        document.getElementById('settings-btn').addEventListener('click', () => {
            alert('設定画面は準備中です');
        });
        
        function updateButtons() {
            document.getElementById('start-btn').disabled = isRunning;
            document.getElementById('stop-btn').disabled = !isRunning;
        }
        
        // Update stats every 2 seconds
        setInterval(updateStats, 2000);
        updateStats();
        
        // Start automatically
        setTimeout(async () => {
            await fetch('http://localhost:8083/start', { method: 'POST' });
            isRunning = true;
            updateButtons();
        }, 1000);
    </script>
</body>
</html>
`