use tokio::runtime::Runtime;
use reqwest::blocking::Client;

pub async fn fetch_data(url: &str) -> Result<String, Box<dyn std::error::Error>> {
    let client = Client::new();
    let response = client.get(url).send().await?;
    let body = response.text().await?;
    Ok(body)
}
