// 测试各种 import 方式
import axios from 'axios';
import { debounce, throttle } from 'lodash';
import React from 'react';
import { createRoot } from 'react-dom/client';
import type { IncomingMessage } from 'http';

// 子路径导入
import { get } from 'lodash/fp';
import axiosRetry from 'axios-retry'; // 未声明的包，应忽略

// 动态 import
const loadMoment = () => import('moment');

// require
const chalk = require('chalk');

export { axios, debounce, throttle, React, createRoot, get, chalk };
