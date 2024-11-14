package com.example;

import org.apache.http.HttpEntity;
import org.apache.http.HttpHeaders;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.apache.http.protocol.HttpContext;
import org.apache.http.protocol.BasicHttpContext;
import org.apache.http.conn.ManagedHttpClientConnection;
import org.apache.http.impl.conn.DefaultManagedHttpClientConnection;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.InetSocketAddress;
import java.net.URI;

public class SimpleCurlClient {

    public static void main(String[] args) throws IOException {
        if (args.length == 0) {
            System.out.println("Usage: java SimpleCurlClient <URL>");
            return;
        }

        URI uri = URI.create(args[0]);
        try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
            BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));

            System.out.println("Press Enter to make the initial request...");
            reader.readLine();
            makeRequest(httpClient, uri);

            System.out.println("Press Enter to make the second request...");
            reader.readLine();
            System.out.println("Making second request to: " + uri);
            makeRequest(httpClient, uri);

            System.out.println("Press Enter to make the third request...");
            reader = new BufferedReader(new InputStreamReader(System.in));
            reader.readLine();
            System.out.println("Making second request to: " + uri);
            makeRequest(httpClient, uri);
        }
    }

    private static void makeRequest(CloseableHttpClient httpClient, URI uri) throws IOException {
        HttpGet request = new HttpGet(uri);
        HttpContext context = new BasicHttpContext();

        try (CloseableHttpResponse response = httpClient.execute(request, context)) {
            System.out.println("Response Status: " + response.getStatusLine());

            String keepAlive = response.getFirstHeader(HttpHeaders.CONNECTION) != null ?
                    response.getFirstHeader(HttpHeaders.CONNECTION).getValue() : "No keep-alive header";
            System.out.println("Connection Header (Keep-Alive): " + keepAlive);

            ManagedHttpClientConnection connection = (ManagedHttpClientConnection) context.getAttribute("http.connection");
            if (connection != null) {
                InetSocketAddress localAddress = (InetSocketAddress) connection.getSocket().getLocalSocketAddress();
                InetSocketAddress remoteAddress = (InetSocketAddress) connection.getSocket().getRemoteSocketAddress();
                System.out.println("Local Address: " + localAddress.getAddress().getHostAddress() + ":" + localAddress.getPort());
                System.out.println("Remote Address: " + remoteAddress.getAddress().getHostAddress() + ":" + remoteAddress.getPort());
            }

            HttpEntity entity = response.getEntity();
            if (entity != null) {
                String responseString = EntityUtils.toString(entity).trim();
                System.out.println("Response Body: ");
                System.out.println(responseString);
            }
        }
    }
}
