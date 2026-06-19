use reqwest::Client;
use tokio::sync::Mutex;
use log::debug;
use std::sync::Arc;

pub struct ApiClient {
    client: Client,
    base_url: String,
}

impl ApiClient {
    pub fn new(base_url: &str) -> Self {
        Self {
            client: Client::new(),
            base_url: base_url.to_string(),
        }
    }
    
    pub async fn get(&self, path: &str) -> Result<String, reqwest::Error> {
        let url = format!("{}/{}", self.base_url, path);
        debug!("GET {}", url);
        self.client.get(&url).send().await?.text().await
    }
}

pub fn create_shared_client() -> Arc<Mutex<ApiClient>> {
    Arc::new(Mutex::new(ApiClient::new("http://localhost:8080")))
}
