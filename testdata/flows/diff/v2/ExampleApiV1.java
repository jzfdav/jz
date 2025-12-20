package test.diff;

import javax.ws.rs.GET;
import javax.ws.rs.Path;
import javax.ws.rs.core.Response;
import javax.ws.rs.client.Client;
import javax.ws.rs.client.ClientBuilder;

@Path("/v1/example")
public class ExampleApiV1 {

    @GET
    public Response handleDiff(String id) {
        if (id == null) {
            return Response.status(400).build();
        }

        if (id.length() < 5) {
            return Response.status(422).build();
        }

        Client client = ClientBuilder.newClient();
        client.target("http://audit-service/v1/log").request().post(null);

        return Response.ok("OK").build();
    }
}
