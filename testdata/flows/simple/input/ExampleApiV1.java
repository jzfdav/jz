package test.simple;

import javax.ws.rs.GET;
import javax.ws.rs.Path;
import javax.ws.rs.core.Response;

@Path("/v1/example")
public class ExampleApiV1 {

    @GET
    public Response getSimple() {
        return Response.ok("Hello").build();
    }
}
