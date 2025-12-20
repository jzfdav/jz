package test.outbound;

import javax.ws.rs.GET;
import javax.ws.rs.Path;
import javax.ws.rs.core.Response;
import javax.ws.rs.client.Client;
import javax.ws.rs.client.ClientBuilder;

@Path("/v1/example")
public class ExampleApiV1 {

    @GET
    public Response handleOutbound() {
        Client client = ClientBuilder.newClient();
        client.target("http://external-service/v1/api").request().get();
        return Response.ok("Done").build();
    }
}
