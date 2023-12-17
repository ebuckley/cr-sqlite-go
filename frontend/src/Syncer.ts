import { DBAsync, StmtAsync, firstPick } from "@vlcn.io/xplat-api";
import {useEffect, useState} from "react";
import {createPromiseClient, Transport, PromiseClient} from '@connectrpc/connect';
import {ChangeService} from "./gen/api/v1/service_connect.ts";

type Args = Readonly<{
  db: DBAsync;
  trans: Transport;
  dbName: string;
  schemaName: string;
  schemaVersion: bigint;
  pullChangesetStmt: StmtAsync;
  applyChangesetStmt: StmtAsync;
  siteId: string;
}>;


class Syncer {
  private client: PromiseClient<typeof ChangeService>;

  constructor(private args: Args) {
    this.client = createPromiseClient(ChangeService, args.trans);
  }
  private lastSentVersionKey(): string {
    return `${this.args.db.siteid}-last-sent-to-backend-${
      this.args.dbName
    }`
  }
  private lastSeenVersionKey(): string {
    return `${this.args.db.siteid}-last-seen-from-backend-${
      this.args.dbName
    }`
  }
  async getLocalChanges(lastSentVersion: BigInt): Promise<any[]> {
    // gather our changes to send to the server
    const changes = await this.args.pullChangesetStmt.all(
      null,
      lastSentVersion
    );
    return changes
  }
  async pushChanges() {
    // track what we last sent to the server so we only send the diff.
    const lastSentVersion = BigInt(
      localStorage.getItem(
        this.lastSentVersionKey()
      ) ?? "0"
    );

    const changes = await this.getLocalChanges(lastSentVersion);

    console.log(`Sending ${changes.length} changes since ${lastSentVersion}`);

    try {
    await this.client.mergeChanges({
      changes: changes.map(c => ({
          table: c[0],
          pk: c[1],
          cid: c[2],
          val: c[3],
          colVersion: c[4],
          dbVersion: c[5],
          siteId: c[6],
          cl: c[7],
          seq: c[8]
        })),
      });
      // Record that we've sent up to the given db version to the server
      // so next sync will be a delta.
      localStorage.setItem(
        this.lastSentVersionKey(),
        changes[changes.length - 1][5].toString(10)
      );
      return changes.length;
    } catch (e) {
      console.error("sync error:", e);
      throw e;
    }

  }

  async pullChanges() {
    const lastSeenVersion = BigInt(
      localStorage.getItem(
        this.lastSeenVersionKey(),
      ) ?? "0"
    );
    const changeRes = await this.client.getChanges({
      dbVersion: lastSeenVersion,
      siteId: this.args.siteId,
    })

    if (changeRes.changes.length == 0) {
      return;
    }
    console.log('Pull has changes to apply: ', changeRes.changes.length, changeRes.changes);
    await this.args.db.tx(async (tx) => {
      for (const c of changeRes.changes) {
        await this.args.applyChangesetStmt.run(
          tx,
          c.table,
          c.pk,
          c.cid,
          c.val,
          c.colVersion,
          c.dbVersion,
          c.siteId,
          c.cl,
          c.seq,
        );
      }
    });

    // Record that we've seen up to the given db version from the server
    // so next sync will be a delta.
    localStorage.setItem(
      this.lastSeenVersionKey(),
      changeRes.changes[changeRes.changes.length - 1].dbVersion.toString(10)
    );

    return changeRes.changes.length;
  }

  destroy() {
    this.args.applyChangesetStmt.finalize(null);
    this.args.pullChangesetStmt.finalize(null);
  }
}

const createSyncer = async (db: DBAsync, dbName: string, trans: Transport): Promise<Syncer> => {
  // @ts-ignore
  const schemaName: string = firstPick<string>(
    await db.execA<[string]>(
      `SELECT value FROM crsql_master WHERE key = 'schema_name'`
    )
  );
  if (schemaName == null) {
    throw new Error("The database does not have a schema applied.");
  } else {
    console.log('schemaName', schemaName);
  }

  const schemaVersion = BigInt(
    firstPick<number | bigint>(
      await db.execA<[number | bigint]>(
        /*sql*/ `SELECT value FROM crsql_master WHERE key = 'schema_version'`
      )
    ) || -1
  );
  console.log('schemaVersion', schemaVersion);

  const [pullChangesetStmt, applyChangesetStmt] = await Promise.all([
    db.prepare(/*sql*/ `
      SELECT "table", "pk", "cid", "val", "col_version", "db_version", "site_id", "cl", "seq" 
      FROM crsql_changes 
      WHERE db_version > ? AND site_id IS NULL
    `),
    db.prepare(/*sql*/ `
      INSERT INTO crsql_changes ("table", "pk", "cid", "val", "col_version", "db_version", "site_id", "cl", "seq")
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `),
  ]);
  pullChangesetStmt.raw(true);
  const siteId = (await db.execA<[string]>(`SELECT quote(crsql_site_id())`))[0][0];


  return new Syncer({
    trans,
    siteId,
    dbName,
    db,
    schemaName,
    schemaVersion,
    applyChangesetStmt,
    pullChangesetStmt
  });
}
export function useSyncer(db: DBAsync, dbName: string, trans: Transport) {
  const [syncer, setSyncer] = useState<Syncer | null>(null);

  useEffect(() => {
    let mounted = true;
    const syncer = createSyncer(db, dbName, trans)
    syncer.then(s => {
      if (!mounted) {
        return
      }
      setSyncer(s);
    })
    return () => {
      if (!mounted) {
        return;
      }
      mounted = false;
      syncer.then(s => s.destroy())
    }
  }, [db, dbName]);

  return syncer
}