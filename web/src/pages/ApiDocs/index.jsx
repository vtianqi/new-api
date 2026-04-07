import React, { useContext } from 'react';
import { Typography, Card, Tag, Tabs, TabPane } from '@douyinfe/semi-ui';
import { StatusContext } from '../../context/Status';

const { Title, Text, Paragraph } = Typography;

const CodeBlock = ({ code, lang = 'bash' }) => (
  <pre style={{
    background: '#1e1e1e', color: '#d4d4d4', padding: '16px',
    borderRadius: '8px', overflow: 'auto', fontSize: '13px',
    lineHeight: '1.6', margin: '8px 0'
  }}>
    <code>{code}</code>
  </pre>
);

const ApiDocs = () => {
  const [statusState] = useContext(StatusContext);
  const baseUrl = window.location.origin;

  const models = [
    { name: 'claude-opus-4-6', desc: '最强推理，复杂任务首选', tier: '高级' },
    { name: 'claude-sonnet-4-6', desc: '性能均衡，日常使用推荐', tier: '标准' },
    { name: 'claude-haiku-4', desc: '速度最快，简单任务首选', tier: '基础' },
  ];

  const curlExample = `curl ${baseUrl}/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-你的APIKey" \\
  -d '{
    "model": "claude-sonnet-4-6",
    "messages": [
      {"role": "user", "content": "你好！"}
    ],
    "stream": false
  }'`;

  const pythonExample = `from openai import OpenAI

client = OpenAI(
    api_key="sk-你的APIKey",
    base_url="${baseUrl}/v1"
)

response = client.chat.completions.create(
    model="claude-sonnet-4-6",
    messages=[
        {"role": "user", "content": "你好！"}
    ]
)
print(response.choices[0].message.content)`;

  const jsExample = `import OpenAI from 'openai';

const client = new OpenAI({
  apiKey: 'sk-你的APIKey',
  baseURL: '${baseUrl}/v1',
});

const response = await client.chat.completions.create({
  model: 'claude-sonnet-4-6',
  messages: [{ role: 'user', content: '你好！' }],
});
console.log(response.choices[0].message.content);`;

  const streamExample = `curl ${baseUrl}/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-你的APIKey" \\
  -d '{
    "model": "claude-sonnet-4-6",
    "messages": [{"role": "user", "content": "写一首诗"}],
    "stream": true
  }'`;

  return (
    <div style={{ maxWidth: 900, margin: '0 auto', padding: '32px 16px' }}>
      <Title heading={2}>API 使用文档</Title>
      <Paragraph type="secondary" style={{ marginBottom: 32 }}>
        本服务兼容 OpenAI API 格式，任何支持自定义 API 地址的客户端均可使用。
      </Paragraph>

      {/* 接入信息 */}
      <Card title="📡 接入信息" style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div>
            <Text strong>API Base URL：</Text>
            <code style={{ background: '#f5f5f5', padding: '2px 8px', borderRadius: 4, marginLeft: 8 }}>
              {baseUrl}/v1
            </code>
          </div>
          <div>
            <Text strong>API Key：</Text>
            <Text type="secondary" style={{ marginLeft: 8 }}>登录后台 → 令牌管理 → 创建令牌</Text>
          </div>
          <div>
            <Text strong>兼容格式：</Text>
            <Tag color="blue" style={{ marginLeft: 8 }}>OpenAI</Tag>
            <Tag color="green" style={{ marginLeft: 4 }}>Claude</Tag>
            <Tag color="orange" style={{ marginLeft: 4 }}>Gemini</Tag>
          </div>
        </div>
      </Card>

      {/* 模型列表 */}
      <Card title="🤖 可用模型" style={{ marginBottom: 24 }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid #eee', textAlign: 'left' }}>
              <th style={{ padding: '8px 12px' }}>模型名称</th>
              <th style={{ padding: '8px 12px' }}>说明</th>
              <th style={{ padding: '8px 12px' }}>档位</th>
            </tr>
          </thead>
          <tbody>
            {models.map(m => (
              <tr key={m.name} style={{ borderBottom: '1px solid #f5f5f5' }}>
                <td style={{ padding: '10px 12px' }}>
                  <code style={{ background: '#f0f0f0', padding: '2px 6px', borderRadius: 3 }}>{m.name}</code>
                </td>
                <td style={{ padding: '10px 12px' }}><Text type="secondary">{m.desc}</Text></td>
                <td style={{ padding: '10px 12px' }}>
                  <Tag color={m.tier === '高级' ? 'red' : m.tier === '标准' ? 'blue' : 'green'} size="small">
                    {m.tier}
                  </Tag>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {/* 代码示例 */}
      <Card title="💻 快速开始" style={{ marginBottom: 24 }}>
        <Tabs type="line">
          <TabPane tab="cURL" itemKey="curl">
            <CodeBlock code={curlExample} lang="bash" />
          </TabPane>
          <TabPane tab="Python" itemKey="python">
            <CodeBlock code={pythonExample} lang="python" />
          </TabPane>
          <TabPane tab="JavaScript" itemKey="js">
            <CodeBlock code={jsExample} lang="javascript" />
          </TabPane>
          <TabPane tab="流式输出" itemKey="stream">
            <Paragraph type="secondary" style={{ marginBottom: 8 }}>
              设置 <code>stream: true</code> 开启流式输出，响应格式为 SSE。
            </Paragraph>
            <CodeBlock code={streamExample} lang="bash" />
          </TabPane>
        </Tabs>
      </Card>

      {/* 常见问题 */}
      <Card title="❓ 常见问题">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div>
            <Text strong>Q: 支持哪些客户端？</Text>
            <Paragraph type="secondary" style={{ margin: '4px 0 0' }}>
              支持所有兼容 OpenAI API 的客户端，包括 Chatbox、Cherry Studio、Open WebUI、Cursor、Copilot 等。
            </Paragraph>
          </div>
          <div>
            <Text strong>Q: 额度不足怎么办？</Text>
            <Paragraph type="secondary" style={{ margin: '4px 0 0' }}>
              登录后台 → 充值中心，支持支付宝、微信、USDT 充值。
            </Paragraph>
          </div>
          <div>
            <Text strong>Q: 请求报错 401 怎么办？</Text>
            <Paragraph type="secondary" style={{ margin: '4px 0 0' }}>
              检查 API Key 是否正确，格式为 <code>Bearer sk-xxx</code>，注意 Bearer 后有空格。
            </Paragraph>
          </div>
          <div>
            <Text strong>Q: 模型降级是什么意思？</Text>
            <Paragraph type="secondary" style={{ margin: '4px 0 0' }}>
              当请求的模型暂时不可用时，系统可能自动切换到同系列低一档的模型。响应头 <code>X-Model-Fallback</code> 会显示实际使用的模型。
            </Paragraph>
          </div>
        </div>
      </Card>
    </div>
  );
};

export default ApiDocs;
