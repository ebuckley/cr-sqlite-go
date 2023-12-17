
import './App.css'
import {createConnectTransport} from "@connectrpc/connect-web";
import {QueryClient, QueryClientProvider} from "@tanstack/react-query";
import {TransportProvider, useQuery} from "@connectrpc/connect-query";
import {useDB, useQuery as useSQL} from "@vlcn.io/react";
import {getSchema, getSiteID} from "./gen/api/v1/service-ChangeService_connectquery.ts";
import {
  GetSchemaRequest,
  GetSchemaResponse,
  GetSiteIDRequest,
  GetSiteIDResponse
} from "./gen/api/v1/service_pb.ts";
// import initWasm, {DB} from "@vlcn.io/crsqlite-wasm";
import {ChangeService} from "./gen/api/v1/service_connect.ts";
import { createPromiseClient } from "@connectrpc/connect";
import {useEffect, useState} from "react";
import { DBProvider } from "@vlcn.io/react";
import {Schema} from "@vlcn.io/react/dist/db/DBFactory";
import {useSyncer} from "./Syncer.ts";

const dbName = 'app.db';

const finalTransport = createConnectTransport({
  baseUrl: "http://localhost:50051",
});


async function init(): Promise<Schema> {
  const client = createPromiseClient(ChangeService, finalTransport);
  // const crsqlite = await initWasm();
  // const db = await crsqlite.open(dbName);
  // const cachedVersion = localStorage.getItem('dbVersion') || -1;
  const spec = await client.getSchema(new GetSchemaRequest());
  return {name: spec.version.toString(), content: spec.schema};
}

// YOINK! https://github.com/ai/nanoid/blob/main/nanoid.js
const nanoid=(t=21)=>crypto.getRandomValues(new Uint8Array(t)).reduce(((t,e)=>t+=(e&=63)<36?e.toString(36):e<62?(e-26).toString(36).toUpperCase():e>62?"-":"_"),"");



function Site() {
  const ctx = useDB(dbName);
  const res = useSQL(ctx, `SELECT * FROM note`)
  const syncer = useSyncer(ctx.db, dbName, finalTransport)

  const pushRow = async () => {
    await ctx.db.execA(`INSERT INTO note (id, title, body) VALUES (?, ?, ?)`, [nanoid(), 'Hello', 'World'])
    await pushChanges()
  }
  useEffect(() => {
    let hadnler: any;
    if (syncer) {
      pullChanges()
      hadnler = setInterval(async () => {
         await pullChanges()
      }, 5_000)

    }
    return () => {
      clearInterval(hadnler)
    }
  }, [syncer]);
  const [pushPullMsg, setPushPullMsg] = useState("");
  const [pushPullTime, setPushPullTime] = useState<Date | null>(null);
  const pushChanges = async () => {
    setPushPullTime(new Date());
    try {
      setPushPullMsg(`Pushing changes...`);
      const num = (await syncer?.pushChanges()) || 0;
      setPushPullMsg(`Pushed ${num} changes`);
    } catch (e: any) {
      setPushPullMsg(`Err pushing: ${e.message}`);
    }
  };
  const pullChanges = async () => {
    setPushPullTime(new Date());
    try {
      setPushPullMsg(`Pulling changes...`);
      console.log('syncer', syncer)
      const num = (await syncer?.pullChanges()) || 0;
      setPushPullMsg(`Pulled ${num} changes`);
    } catch (e: any) {
      console.log(e);
      setPushPullMsg(`Err pulling: ${e.message || e}`);
    }
  };

  // const res = useSQL(`SELECT * FROM note`);
  // @ts-ignore
  const schemaQuery = useQuery<GetSchemaRequest, GetSchemaResponse>(getSchema, {})
  // @ts-ignore
  const siteQuery = useQuery<GetSiteIDRequest, GetSiteIDResponse>(getSiteID)

  const getLoc = async () => {
    const changes = await syncer?.getLocalChanges(BigInt(0));
    console.log('changes', changes)
  }
  return <div className="App">
    <h2>Results of a query!!!</h2>
    <pre>{res.loading || JSON.stringify(res.data, null, '  ')}</pre>
    <button onClick={pushRow}>Push in to table...</button>

    <button onClick={getLoc}>Get local Changes</button>
    <div>Last Pulled: <span style={{display: 'inline-block', width: '92px', background: '#a2a2a2'}}>{pushPullTime?.toLocaleTimeString()}</span>
    </div>
    <div>{'' || pushPullMsg}</div>
    <details>
      <summary>
        Remote Database Schema
      </summary>
      <dl>
        <dt>Schema</dt>
        <dd>{schemaQuery.data && schemaQuery.data.schema}</dd>
        <dt>Version</dt>
        <dd>{schemaQuery.data && schemaQuery.data.version.toString()}</dd>
        <dt>Remote Site ID</dt>
        <dd>{siteQuery.data && siteQuery.data.siteId}</dd>
      </dl>
    </details>
  </div>
}

const queryClient = new QueryClient();

function App() {
  const [schema, setDb] = useState<Schema|undefined>();
  useEffect(() => {
    init().then((db) => {
      setDb(db);
    })
  }, []);

  return schema &&
    <DBProvider dbname={dbName} schema={schema && schema} Render={ ()=>
      <TransportProvider transport={finalTransport}>
        <QueryClientProvider client={queryClient}>
          <Site/>
        </QueryClientProvider>
      </TransportProvider>}/>
}

export default App
