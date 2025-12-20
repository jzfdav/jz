package test.diff;

import javax.ws.rs.GET;
import javax.ws.rs.Path;
import javax.ws.rs.core.Response;

@Path("/v1/example")
public class ExampleApiV1 {

    @GET
    public Response handleDiff(String id) {
        if (id == null) {
            return Response.status(400).build();
        }
        return Response.ok("OK").build();
    }
}
