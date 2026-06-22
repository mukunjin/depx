import express from 'express';
import { merge } from 'lodash';
import { jest } from '@jest/globals';

const app = express();
const config = merge({ a: 1 }, { b: 2 });

jest.mock('lodash');

export { app, config };
