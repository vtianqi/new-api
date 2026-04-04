import React, { useState, useEffect } from 'react';
import { Card, Col, Row, Select, Spin, Typography, Table } from '@douyinfe/semi-ui';
import { API } from '../../helpers';

const { Title, Text } = Typography;

const RevenueUser = () => {
  const [loading, setLoading] = useState(false);
  const [days, setDays] = useState(7);
  const [data, setData] = useState(null);

  const fetchData = async (d) => {
    setLoading(true);
    try {
      const res = await API.get(`/api/log/revenue/self?days=${d}`);
      if (res.data.success) setData(res.data.data);
    } catch (e) {}
    setLoading(false);
  };

  useEffect(() => { fetchData(days); }, [days]);

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16, gap: 16 }}>
        <Title heading={4} style={{ margin: 0 }}>我的用量统计</Title>
        <Select value={days} onChange={v => { setDays(v); fetchData(v); }} style={{ width: 120 }}>
          <Select.Option value={7}>近7天</Select.Option>
          <Select.Option value={14}>近14天</Select.Option>
          <Select.Option value={30}>近30天</Select.Option>
        </Select>
      </div>

      <Spin spinning={loading}>
        {data && (
          <>
            <Row gutter={16} style={{ marginBottom: 24 }}>
              <Col span={8}>
                <Card><Text type="secondary">总消费</Text><Title heading={3}>¥{data.summary.total_cost.toFixed(4)}</Title></Card>
              </Col>
              <Col span={8}>
                <Card><Text type="secondary">总调用次数</Text><Title heading={3}>{data.summary.total_calls}</Title></Card>
              </Col>
              <Col span={8}>
                <Card><Text type="secondary">消耗Quota</Text><Title heading={3}>{data.summary.total_quota}</Title></Card>
              </Col>
            </Row>

            <Card title="每日用量" style={{ marginBottom: 24 }}>
              <Table
                dataSource={data.daily}
                columns={[
                  { title: '日期', dataIndex: 'date' },
                  { title: '消费(元)', dataIndex: 'cost', render: v => v.toFixed(4) },
                  { title: '调用次数', dataIndex: 'calls' },
                  { title: 'Quota', dataIndex: 'quota' },
                ]}
                pagination={false}
                size="small"
              />
            </Card>

            <Card title="模型使用分布">
              <Table
                dataSource={data.models}
                columns={[
                  { title: '模型', dataIndex: 'model_name' },
                  { title: '调用次数', dataIndex: 'calls' },
                  { title: 'Quota', dataIndex: 'quota' },
                ]}
                pagination={false}
                size="small"
              />
            </Card>
          </>
        )}
      </Spin>
    </div>
  );
};

export default RevenueUser;
