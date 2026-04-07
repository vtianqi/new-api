import React, { useState, useEffect } from 'react';
import { Card, Button, Input, TextArea, Tag, Typography, Modal, Form, Select, Spin } from '@douyinfe/semi-ui';
import { API } from '../../helpers';

const { Title, Text, Paragraph } = Typography;

const statusMap = { 1: { label: '待处理', color: 'orange' }, 2: { label: '处理中', color: 'blue' }, 3: { label: '已解决', color: 'green' }, 4: { label: '已关闭', color: 'grey' } };

const TicketPage = () => {
  const [tickets, setTickets] = useState([]);
  const [loading, setLoading] = useState(false);
  const [createVisible, setCreateVisible] = useState(false);
  const [detail, setDetail] = useState(null);
  const [replyContent, setReplyContent] = useState('');

  const fetchTickets = async () => {
    setLoading(true);
    const res = await API.get('/api/ticket/');
    if (res.data.success) setTickets(res.data.data || []);
    setLoading(false);
  };

  const fetchDetail = async (id) => {
    const res = await API.get(`/api/ticket/${id}`);
    if (res.data.success) setDetail(res.data.data);
  };

  const createTicket = async (values) => {
    const res = await API.post('/api/ticket/', values);
    if (res.data.success) { setCreateVisible(false); fetchTickets(); }
  };

  const sendReply = async () => {
    if (!replyContent.trim() || !detail) return;
    await API.post(`/api/ticket/${detail.ticket.id}/reply`, { content: replyContent });
    setReplyContent('');
    fetchDetail(detail.ticket.id);
  };

  useEffect(() => { fetchTickets(); }, []);

  if (detail) {
    const { ticket, messages } = detail;
    return (
      <div style={{ maxWidth: 800, margin: '0 auto', padding: 24 }}>
        <Button onClick={() => setDetail(null)} style={{ marginBottom: 16 }}>← 返回列表</Button>
        <Card title={<div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span>{ticket.title}</span>
          <Tag color={statusMap[ticket.status]?.color}>{statusMap[ticket.status]?.label}</Tag>
        </div>}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12, maxHeight: 400, overflowY: 'auto', marginBottom: 16 }}>
            {messages?.map(m => (
              <div key={m.id} style={{ display: 'flex', justifyContent: m.is_admin ? 'flex-start' : 'flex-end' }}>
                <div style={{
                  maxWidth: '70%', padding: '10px 14px', borderRadius: 8,
                  background: m.is_admin ? '#f5f5f5' : '#e3f2fd',
                }}>
                  <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
                    {m.is_admin ? '客服' : m.username} · {new Date(m.created_time * 1000).toLocaleString()}
                  </Text>
                  <Paragraph style={{ margin: 0 }}>{m.content}</Paragraph>
                </div>
              </div>
            ))}
          </div>
          {ticket.status < 3 && (
            <div style={{ display: 'flex', gap: 8 }}>
              <Input value={replyContent} onChange={setReplyContent} placeholder="输入回复内容..." onEnterPress={sendReply} />
              <Button type="primary" onClick={sendReply}>发送</Button>
            </div>
          )}
        </Card>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: 800, margin: '0 auto', padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title heading={4} style={{ margin: 0 }}>我的工单</Title>
        <Button type="primary" onClick={() => setCreateVisible(true)}>提交工单</Button>
      </div>
      <Spin spinning={loading}>
        {tickets.length === 0 ? (
          <Card><Text type="secondary">暂无工单</Text></Card>
        ) : tickets.map(t => (
          <Card key={t.id} style={{ marginBottom: 12, cursor: 'pointer' }} onClick={() => fetchDetail(t.id)}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <Text strong>{t.title}</Text>
                <Text type="secondary" style={{ marginLeft: 8, fontSize: 12 }}>
                  {new Date(t.updated_time * 1000).toLocaleString()}
                </Text>
              </div>
              <Tag color={statusMap[t.status]?.color}>{statusMap[t.status]?.label}</Tag>
            </div>
          </Card>
        ))}
      </Spin>

      <Modal title="提交工单" visible={createVisible} onCancel={() => setCreateVisible(false)} footer={null}>
        <Form onSubmit={createTicket}>
          <Form.Input field="title" label="标题" rules={[{ required: true, message: '请输入标题' }]} />
          <Form.Select field="priority" label="优先级" initValue={1}>
            <Select.Option value={1}>普通</Select.Option>
            <Select.Option value={2}>紧急</Select.Option>
          </Form.Select>
          <Form.TextArea field="content" label="描述问题" rules={[{ required: true, message: '请描述问题' }]} rows={4} />
          <Button htmlType="submit" type="primary" block style={{ marginTop: 8 }}>提交</Button>
        </Form>
      </Modal>
    </div>
  );
};

export default TicketPage;
