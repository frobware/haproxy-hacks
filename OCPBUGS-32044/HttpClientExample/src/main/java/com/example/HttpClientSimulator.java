package com.example;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import org.apache.http.NoHttpResponseException;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.DefaultHttpRequestRetryHandler;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.impl.conn.PoolingHttpClientConnectionManager;

public class HttpClientSimulator {
        public static void main(String[] args) {
                if (args.length < 1) {
                        System.out.println("Usage: <URL>");
                        System.exit(1);
                }

                String url = args[0];
                PoolingHttpClientConnectionManager cm = new PoolingHttpClientConnectionManager();

                // Set max total connections
                cm.setMaxTotal(200);

                // Set max connections per route
                cm.setDefaultMaxPerRoute(100);

                HttpClientBuilder builder = HttpClients.custom();

                // Set connection manager.
                builder.setConnectionManager(cm);

                // Set retry handler.
                builder.setRetryHandler(new DefaultHttpRequestRetryHandler(2, false));

                // Build CloseableHttpClient.
                CloseableHttpClient httpClient = builder.build();

                ExecutorService executor = Executors.newFixedThreadPool(100);

                // Simulate multiple clients making requests.
                Runnable task = () -> {
                        while (!Thread.currentThread().isInterrupted()) {
                                try (CloseableHttpResponse response = httpClient.execute(new HttpGet(url))) {
                                        System.out.println("Response: \"" + response.getStatusLine() + "\" Status: " + response.getStatusLine().getStatusCode());
                                        TimeUnit.MILLISECONDS.sleep(1);
                                } catch (NoHttpResponseException e) {
                                        System.err.println("NO RESPONSE EXCEPTION FROM SERVER, LIKELY DUE TO CONNECTION RESET: " + e.getMessage());
                                } catch (Exception e) {
                                        System.err.println("Error during request: " + e.getClass().getSimpleName());
                                        e.printStackTrace();
                                }
                        }
                };

                for (int i = 0; i < 100; i++) {
                        executor.submit(task);
                }

                Runtime.getRuntime().addShutdownHook(new Thread(() -> {
                                        executor.shutdownNow();
                                        try {
                                                httpClient.close();
                                        } catch (Exception e) {
                                                System.err.println("Error closing client: " + e.getMessage());
                                        }
                }));

                try {
                        executor.awaitTermination(Long.MAX_VALUE, TimeUnit.DAYS);
                } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                }
        }
}
