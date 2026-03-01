const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

app.get('/', (req, res) => {
  res.json({
    message: '🚀 デプロイ成功！',
    app: 'deploy-test',
    timestamp: new Date().toISOString(),
    env: {
      node: process.version,
      platform: process.platform,
      uptime: Math.floor(process.uptime()) + 's'
    }
  });
});

app.get('/hello', (req, res) => {
  const name = req.query.name || 'World';
  res.json({ greeting: `Hello, ${name}!` });
});

app.get('/health', (req, res) => {
  res.json({ status: 'ok', uptime: process.uptime() });
});

app.listen(PORT, () => {
  console.log(`✅ Server running on port ${PORT}`);
});
