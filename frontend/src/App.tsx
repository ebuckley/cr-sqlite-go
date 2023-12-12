
import './App.css'
import {createConnectTransport} from "@connectrpc/connect-web";
import {QueryClient, QueryClientProvider} from "@tanstack/react-query";
import {TransportProvider, useQuery} from "@connectrpc/connect-query";
import {getSiteID} from "./gen/api/v1/service-ChangeService_connectquery.ts";
import {GetSiteIDRequest, GetSiteIDResponse} from "./gen/api/v1/service_pb.ts";


const finalTransport = createConnectTransport({
  baseUrl: "http://localhost:50051",
});

function Site() {

  // @ts-ignore
  const siteQuery = useQuery<GetSiteIDRequest, GetSiteIDResponse>(getSiteID)
  return <div className="App">
    <h2>Remote Site ID</h2>
    <pre>{siteQuery.data && JSON.stringify(siteQuery.data, null, '  ')}</pre>
  </div>
}

const queryClient = new QueryClient();

function App() {

  return (
    <TransportProvider transport={finalTransport}>
      <QueryClientProvider client={queryClient}>
        <Site/>
      </QueryClientProvider>
    </TransportProvider>
  )
}

export default App
