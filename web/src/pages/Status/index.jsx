import React, { useState, useEffect } from 'react';
import { Card, Typography, Tag, Spin } from '@douyinfe/semi-ui';

const { Title, Text } = Typography;

const StatusPage = () => {
  const [loading, setLoading] = useState(true);
  const [apiStatus, setApiStatus] = useState(null);
  const [lastCheck, setLastCheck] = useState(null);

  const checkStatus = async () => {
    setLoading(true);
    try {
      const start = Date.now();
      const res = await fetch('/api/status', { method: 'GET' });
      const latency = Date.now() - start;
      const data = await res.json();
      setApiStatus({ ok: res.ok && data.success !== false, latency, data });
    } catch (e) {
      setApiStatus({ ok: false, latency: null, error: e.message });
    }
    setLastCheck(new Date());
    setLoading(false);
  };

  useEffect(() => {
    checkStatus();
    const timer = setInterval(checkStatus, 30000);
    return () => clearInterval(timer);
  }, []);

  const services = [
    { name: 'API 服务', status: apiStatus?.ok, latency: apiStatus?.latency },
    { name: 'Claude 模型', status: apiStatus?.ok ? true : false },
    { name: '计费系统', status: apiStatus?.ok ? true : false },
  ];

  const allOk = services.every(s => s.status);

  return (
    <div style={{ maxWidth: 700, margin: '0 auto', padding: '40px 16px' }}>
      <div style={{ textAlign: 'center', marginBottom: 32 }}>
        <Title heading={2}>系统状态</Title>
        <div style={{
          display: 'inline-flex', alignItems: 'center', gap: 8,
          background: allOk ? '#e8f5e9' : '#ffebee',
          padding: '12px 24px', borderRadius: 8, marginTop: 16
        }}>
          <div style={{
            width: 12, height: 12, borderRadius: '50%',
            background: allOk ? '#4caf50' : '#f44336',
            animation: 'pulse 2s infinite'
          }} />
          <Text strong style={{ color: allOk ? '#2e7d32' : '#c62828' }}>
            {loading ? '检测中...' : allOk ? '所有系统正常运行' : '服务异常，正在处理'}
          </Text>
        </div>
      </div>

      <Spin spinning={loading}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {services.map(s => (
            <Card key={s.name} style={{ padding: '4px 0' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Text strong>{s.name}</Text>
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  {s.latency && (
                    <Text type="secondary" style={{ fontSize: 13 }}>{s.latency}ms</Text>
                  )}
                  <Tag color={s.status ? 'green' : 'red'} size="large">
                    {s.status === undefined ? '检测中' : s.status ? '正常' : '异常'}
                  </Tag>
                </div>
              </div>
            </Card>
          ))}
        </div>
      </Spin>

      {lastCheck && (
        <Text type="secondary" style={{ display: 'block', textAlign: 'center', marginTop: 24, fontSize: 13 }}>
          最后检测时间：{lastCheck.toLocaleTimeString()} · 每30秒自动刷新
        </Text>
      )}

      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }
      `}</style>
    </div>
  );
};

export default StatusPage;
