# WAL

Build a simple write-ahead log storage in Go.

> [!Note]
>
> This project is started for an educational purposes to learn more about internals of databases and to get the hands dirty.

## Introduction to WAL

- Write-Ahead-Log(WAL) is an essential component for the durability for databases.
- WALs are more generally used in database systems to flush all written data to disk(non-volatile storage) before the changes are written to the database. So that in case of a crash, lost in-memory changes of database can be recovered using logs.
- WAL file can be read during the restart and state can be recovered just by replaying all the log entries.

<p align="left">
  <img src="./docs/images/basic_archi.png" height="320" width="550" title="Basic WAL architecture">
</p>

## More Details about WAL

- WAL is more than just a log entries example: WAL is used in LSM trees for reliability. Moreover, [RAFT](https://en.wikipedia.org/wiki/Raft_(algorithm)) algorithm uses replicated WALs to achieve consensus in distributed system with the help of log replication.
- Generally WAL files uses append-only file technique, which is basically to append every new entry to the end of the current file.
- The main disadvantage of this append only method is that it will keep growing as data cannot be modified or deleted, hence it can lead to a large number of files, which can make it difficult to manage. To solve such issues techiniques like `Segmented Log` and `Low-Water-Mark` are used.

  - **Segmented Log**: A single log file can grow and become very large which makes periodic cleanup operation difficult, so single log file is split into multiple segments. When the WAL file reaches a configurable maximum size, it is closed and the WAL starts to append its records to a new and empty file. These files are called WAL  *segments.*
  - **Low-Water Mark**: It's lowest offset/index or low-water mark in WAL to tell the logging mechanism which portion of log can be safely discarded.

### Implementation considerations

- **Write operations:** Flushing the logs to disk(non-volatile storage) is done using [fsync](https://en.wikipedia.org/wiki/Fsync). But there is a tradeoff:
  - Flushing every log write to disk gives very strong durability but this limits the write performance. If flushing is done periodically (with a set-interval), it improves the write performance but there is a risk of losing entries (if server crashed before flushing).
  - [Fsync](https://en.wikipedia.org/wiki/Fsync): OS also use their own buffers for disk file operations, which means data written to disk might initially only be stored in an OS buffer and can be lost in case of power failure. To address this, operating systems provide an fsync API to force the synchronization of the OS buffer with the disk. However, fsync can slow down write operations.
  - So, there always will be a tradeoff between write operations performance and durability. So, we need to have a balance b/w both depends on the usecase.
- **Serialization**: Data written to disk must be serialize. Many systems use [Google protobuf](https://protobuf.dev/) for this purpose, which provides rapid serialization and efficient encoding.
- **Concurrency**: WAL file can be opened either in read mode or write(append) mode but not both. Simultaneous read and write operations are not permitted as it will never be required because WAL file are only read during start-up.
- **Integrity**: To ensure the integrity of entries, each log entry should include a checksum which can be verified during the file read and can detect the corrupted entries.

## WAL File Format

- Here, I'm simply taking the simple wal file format. This diagram shows detailed format of the WAL file consider in this project
- Each WAL file consists of a series of records. It’s encoded in little endian.
- The first 4 bytes(32bit) in each record is the log sequence number(offset), in order to implement the low-water mark.
- Then next field is used for CRC, 4bytes(32bit) CRC32 checksum of record data.
- Last field will contain the actual payload.

<p align="left">
  <img src="./docs/images/wal_file_format.png" height="320" width="550" title="Basic WAL architecture">
</p>

- WAL file format can contain more information which is essential to recover the database, take a look at [etcd wal file](https://github.com/ahrtr/etcd-issues/blob/master/docs/cncf_storage_tag_etcd.md#storage-wal-file-format).

## Features

- [ ] Efficient and safe Write-Ahead Log for databases
- [ ] Log Segmentation
- [ ] Integrity check
- [ ] Repair of corrupt wal file

## References

- [https://www.postgresql.org/docs/9.1/wal-intro.html](https://www.postgresql.org/docs/9.1/wal-intro.html)
- [Patterns of Distributed Systems](https://martinfowler.com/books/patterns-distributed.html)
- [Etcd&#39;s wal implementation](https://github.com/liqingqiya/readcode-etcd-v3.4.10/blob/master/src/go.etcd.io/etcd/wal/decoder.go#L35)

## Contributing

I'm sharing this as a starting point for others to learn more about internals of database, and also to gather feedback to learn more. If you have a suggestion, feel free to [file an issue](https://github.com/ishan16696/wal/issues).

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/ishan16696/wal/blob/main/LICENSE) file for details.
