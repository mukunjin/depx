// 嵌套目录中的文件
import axios from 'axios';
import { useState } from 'react';

export function useApi() {
  const [data, setData] = useState(null);
  return axios.get('/api').then(res => setData(res.data));
}
