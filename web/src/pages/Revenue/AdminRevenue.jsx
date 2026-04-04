import React, { useState, useEffect } from 'react';
import { Card, Col, Row, Select, Spin, Typography, Table } from '@douyinfe/semi-ui';
import { API } from '../../helpers';

const { Title, Text } = Typography;

const RevenueAdmin = () => {
  const [loading, setLoading] = useState(false);
  const [days, setDays] = useState(7);
  const [data, setData] = useState(null);

  const fetchData = async (d) => {
    setLoading(true);
    try {
      const res = await API.get(`/api/log/revenue?days=${d}`);
      if (res.data.success) setData(res.data.data);
    } catch (e) {}
    setLoading(false);
  };

  useEffect(() => { fetchData(days); }, [days]);

  const modelColumns = [
    { title: '模型', dataIndex: 'model', key: 'model' },
    { title: '调用次数', dataIndex: 'calls', key: 'calls' },
    { title: '消耗Quota', dataIndex: 'quota', key: 'quota' },
    { title: '营收(元)', dataIndex: 'revenue', key: 'revenue', render: v => v.toFixed(4) },
    { title: '占比', dataIndex: 'percent', key: 'percent', render: v => `${v.toFixed(1)}%` },
  ];

  const userColumns = [
    { title: '用户', dataIndex: 'username', key: 'username' },
    { title: '调用次数', dataIndex: 'calls', key: 'calls' },
    { title: '消耗Quota', dataIndex: 'quota', key: 'quota' },
    { title: '消费(元)', dataIndex: 'revenue', key: 'revenue', render: v => v.toFixed(4) },
  ];

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16, gap: 16 }}>
        <Title heading={4} style={{ margin: 0 }}>营收分析</Title>
        <Select value={days} onChange={v => { setDays(v); fetchData(v); }} style={{ width: 120 }}>
          <Select.Option value={7}>近7天</Select.Option>
          <Select.Option value={14}>近14天</Select.Option>
          <Select.Option value={30}>近30天</Select.Option>
          <Select.Option value={90}>近90天</Select.Option>
        </Select>
      </div>

      <Spin spinning={loading}>
        {data && (
          <>
            {/* 汇总卡片 */}
            <Row gutter={16} style={{ marginBottom: 24 }}>
              <Col span={6}>
                <Card><Text type="secondary">总营收</Text><Title heading={3}>¥{data.summary.total_revenue.toFixed(2)}</Title></Card>
              </Col>
              <Col span={6}>
                <Card><Text type="secondary">总调用次数</Text><Title heading={3}>{data.summary.total_calls}</Title></Card>
              </Col>
              <Col span={6}>
                <Card><Text type="secondary">活跃用户数</Text><Title heading={3}>{data.summary.total_users}</Title></Card>
              </Col>
              <Col span={6}>
                <Card><Text type="secondary">总消耗Quota</Text><Title heading={3}>{data.summary.total_quota}</Title></Card>
              </Col>
            </Row>

            {/* 每日营收表 */}
            <Card title="每日营收趋势" style={{ marginBottom: 24 }}>
              <Table
                dataSource={data.daily}
                columns={[
                  { title: '日期', dataIndex: 'date', key: 'date' },
                  { title: '营收(元)', dataIndex: 'revenue', key: 'revenue', render: v => v.toFixed(4) },
                  { title: '调用次数', dataIndex: 'calls', key: 'calls' },
                  { title: '活跃用户', dataIndex: 'users', key: 'users' },
                  { title: 'Quota', dataIndex: 'quota', key: 'quota' },
                ]}
                pagination={false}
                size="small"
              />
            </Card>

            <Row gutter={16}>
              {/* 模型排行 */}
              <Col span={12}>
                <Card title="热门模型排行">
                  <Table dataSource={data.models} columns={modelColumns} pagination={false} size="small" />
                </Card>
              </Col>
              {/* 用户排行 */}
              <Col span={12}>
                <Card title="活跃用户排行">
                  <Table dataSource={data.top_users} columns={userColumns} pagination={false} size="small" />
                </Card>
              </Col>
            </Row>
          </>
        )}
      </Spin>
    </div>
  );
};

export default RevenueAdmin;
