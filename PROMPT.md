# Hyttekos

I want you to build a webapp for managing the state of heating on/off and hot water on/off for a cabin, as well as reporting the current temperature in the cabin.

The app should be built using a backend and a frontend. 

## Backend

### Integration with Akiles

To manage the heating on/off and hot water on/off, use the Akiles API available at https://dev.akiles.app/api/reference/ - A client id and client secret will be provided as environment variables
to the backend in order to handle the oauth login, redirect and auth.

In addition to authentication, the backend will be configured to use a specific organization and gadgets for the heating and hot water (also via env vars). 

The backend will use gadget action API to trigger on/off of each gadget https://dev.akiles.app/api/reference/#/components/schemas/gadget_action . The get gadget API will be used to fetch the current state of the gadget 'state_id' to determine if its on/off.


### API

The backend should provide an API for the frontend to get/set the current state of heating and hot water, as well as reading the current temperature.

It should also provide an API for setting the temperature (Not authenticated by oauth, but authenticated with a simple HTTP plain auth that is provided as env to the backend).

### Implementation

The implementation should be written in Golang, and should persist the temperature readings. The heating/hot water state is stored in Akiles, so no need to store that in the hyttekos backend.

## Frontend

### Integration with backend

The frontend should provide a single page that shows the current state of heating on/off, hot water on/off and current temperature. The frontend should authenticate itself with the backend using oauth, and should only use the backend API (not go directly to akiles).

### Design

The frontend should use a cozy design to get that cosy cabin feeling. 


### Implementation

The frontend should use minimal javascript, and no framework other than htmx.

## Summary

Before proceeding to implementation, provide detailed design documents of the architecture, design sketches, API and data flows, and a list of tasks for implementing the system.

Write all documents and mockups in separate markdown documents, and generate graphical mockups for the UI as svg.
