import { transform } from '@babel/core';
import presetEnv from '@babel/preset-env';
import type { Config } from '@types/node';
import { parse } from '@typescript-eslint/parser';
import { cloneDeep } from 'lodash-es';
import { createRoot } from 'react-dom/client';

const root = createRoot(document.getElementById('app'));
const config = {} as Config;
const result = transform('code', { presets: [presetEnv] });
const cloned = cloneDeep({ a: 1 });
