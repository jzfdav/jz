package test.guards;

import javax.ws.rs.POST;
import javax.ws.rs.Path;
import javax.ws.rs.core.Response;

@Path("/v1/example")
public class ExampleApiV1 {

    @POST
    public Response handleGuards(String input) {
        if (input == null) {
            return Response.status(400).build();
        }
        
        if (input.isEmpty()) {
            return Response.status(422).build();
        }

        return Response.ok("Valid").build();
    }
}
