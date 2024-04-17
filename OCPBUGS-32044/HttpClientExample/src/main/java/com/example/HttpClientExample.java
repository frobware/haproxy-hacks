package com.example;

import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;

public class HttpClientExample {
    public static void main(String[] args) {
        // Create an HttpClient object
        try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
            // Create an HttpGet object with the URL to send the GET request
            HttpGet request = new HttpGet("http://example.com");

            // Execute the request and receive the response
            try (CloseableHttpResponse response = httpClient.execute(request)) {
                // Convert the response entity to a string
                String result = EntityUtils.toString(response.getEntity());

                // Output the result
                System.out.println(result);
            }
        } catch (Exception e) {
            // Handle exceptions
            e.printStackTrace();
        }
    }
}
