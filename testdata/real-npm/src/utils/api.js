const axios = require('axios');
const debug = require('debug')('app:api');

class ApiClient {
  constructor(baseURL) {
    this.client = axios.create({ baseURL });
  }

  async get(endpoint) {
    debug(`GET ${endpoint}`);
    return this.client.get(endpoint);
  }
}

module.exports = ApiClient;
