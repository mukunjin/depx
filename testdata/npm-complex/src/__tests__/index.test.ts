// 测试文件 - 验证测试文件不会被扫描
import { something } from 'unused-in-test';
import axios from 'axios';

describe('test', () => {
  it('should work', () => {
    expect(true).toBe(true);
  });
});
