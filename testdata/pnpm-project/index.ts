import Fastify from 'fastify';
import { z } from 'zod';

const fastify = Fastify({ logger: true });

const schema = z.object({
  name: z.string(),
  age: z.number()
});

fastify.get('/', async () => {
  return { hello: 'world' };
});

export { fastify, schema };
