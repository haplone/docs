



tidb-lightning 将gc时间设置为100h了，这样当时就不用gc，加速倒数


```log
[2019/08/13 14:54:45.651 +08:00] [INFO] [gc_worker.go:249] ["[gc worker] starts the whole job"] [uuid=5b2140c368c0020] [safePoint=410317589203124224]
[2019/08/13 14:54:45.652 +08:00] [INFO] [gc_worker.go:548] ["[gc worker] start resolve locks"] [uuid=5b2140c368c0020] [safePoint=410317589203124224]






[2019/07/04 10:44:43.451 +08:00] [INFO] [gc_worker.go:1023] ["we change"] [key=tikv_gc_run_interval] [value=200m0s]
[2019/07/04 10:44:43.465 +08:00] [INFO] [gc_worker.go:250] ["[gc worker] starts the whole job"] [uuid=5aec51ce17c0132] [safePoint=409429156408066048]
[2019/07/04 10:44:43.465 +08:00] [INFO] [gc_worker.go:549] ["[gc worker] start resolve locks"] [uuid=5aec51ce17c0132] [safePoint=409429156408066048]


[2019/07/04 11:26:16.780 +08:00] [INFO] [lock_resolver.go:197] ["BatchResolveLocks: lookup txn status"] ["cost time"=48.572269ms] ["num of txn"=1]
[2019/07/04 11:26:16.789 +08:00] [INFO] [lock_resolver.go:242] ["BatchResolveLocks: resolve locks in a batch"] ["cost time"=9.524497ms] ["num of locks"=557]
[2019/07/04 11:26:16.942 +08:00] [INFO] [lock_resolver.go:197] ["BatchResolveLocks: lookup txn status"] ["cost time"=2.893µs] ["num of txn"=1]
[2019/07/04 11:26:16.944 +08:00] [INFO] [lock_resolver.go:242] ["BatchResolveLocks: resolve locks in a batch"] ["cost time"=2.053711ms] ["num of locks"=23]
[2019/07/04 11:26:16.951 +08:00] [INFO] [lock_resolver.go:197] ["BatchResolveLocks: lookup txn status"] ["cost time"=3.346µs] ["num of txn"=1]
[2019/07/04 11:26:16.953 +08:00] [INFO] [lock_resolver.go:242] ["BatchResolveLocks: resolve locks in a batch"] ["cost time"=1.857225ms] ["num of locks"=69]
[2019/07/04 11:26:16.956 +08:00] [INFO] [lock_resolver.go:197] ["BatchResolveLocks: lookup txn status"] ["cost time"=5.036µs] ["num of txn"=1]
[2019/07/04 11:26:16.960 +08:00] [INFO] [lock_resolver.go:242] ["BatchResolveLocks: resolve locks in a batch"] ["cost time"=3.860958ms] ["num of locks"=184]
[2019/07/04 11:26:16.962 +08:00] [INFO] [lock_resolver.go:197] ["BatchResolveLocks: lookup txn status"] ["cost time"=5.51µs] ["num of txn"=1]
[2019/07/04 11:26:16.966 +08:00] [INFO] [lock_resolver.go:242] ["BatchResolveLocks: resolve locks in a batch"] ["cost time"=4.272504ms] ["num of locks"=208]


1342 /data/deploy/log/tidb-2019-07-09T05-05-57.899.log:[2019/07/04 10:44:43.465 +08:00] [INFO] [gc_worker.go:549] ["[gc      worker] start resolve locks"] [uuid=5aec51ce17c0132] [safePoint=409429156408066048]
1387 /data/deploy/log/tidb-2019-07-09T05-05-57.899.log:[2019/07/04 11:28:48.865 +08:00] [INFO] [gc_worker.go:626] ["[gc      worker] finish resolve locks"] [uuid=5aec51ce17c0132] [safePoint=409429156408066048] [regions=437245] ["total res     olved"=3557] ["cost time"=44m5.400141913s]


[2019/08/13 10:08:44.455 +08:00] [INFO] [gc_worker.go:1023] ["we change"] [key=tikv_gc_run_interval] [value=200m0s]
[2019/08/13 10:08:44.463 +08:00] [INFO] [gc_worker.go:328] ["[gc worker] last safe point is later than current one.No need to gc.This might be caused by manually enlarging gc lifetime"] ["leaderTick on"=5b0a56f99bc008f] ["last safe point"=2019/08/08 12:05:44.000 +08:00] ["current safe point"=2019/08/08 06:08:44.421 +08:00]



[2019/08/13 10:31:44.723 +08:00] [INFO] [gc_worker.go:328] ["[gc worker] last safe point is later than current one.No need to gc.This might be caused by manually enlarging gc lifetime"] ["leaderTick on"=5b0a56f99bc008f] ["last safe point"=2019/08/08 12:05:44.000 +08:00] ["current safe point"=2019/08/08 07:31:44.671 +08:00]
```



```rust
// storage/mvvc/lock.rs
impl Lock {
    pub fn new(
        lock_type: LockType,
        primary: Vec<u8>,
        ts: u64,
        ttl: u64,
        short_value: Option<Value>,
    ) -> Lock {
        Lock {
            lock_type,
            primary,
            ts,
            ttl,
            short_value,
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut b = Vec::with_capacity(
            1 + MAX_VAR_U64_LEN + self.primary.len() + MAX_VAR_U64_LEN + SHORT_VALUE_MAX_LEN + 2,
        );
        b.push(self.lock_type.to_u8());
        b.encode_compact_bytes(&self.primary).unwrap();
        b.encode_var_u64(self.ts).unwrap();
        b.encode_var_u64(self.ttl).unwrap();
        if let Some(ref v) = self.short_value {
            b.push(SHORT_VALUE_PREFIX);
            b.push(v.len() as u8);
            b.extend_from_slice(v);
        }
        b
    }

    pub fn parse(mut b: &[u8]) -> Result<Lock> {
        if b.is_empty() {
            return Err(Error::BadFormatLock);
        }
        let lock_type = LockType::from_u8(b.read_u8()?).ok_or(Error::BadFormatLock)?;
        let primary = bytes::decode_compact_bytes(&mut b)?;
        let ts = number::decode_var_u64(&mut b)?;
        let ttl = if b.is_empty() {
            0
        } else {
            number::decode_var_u64(&mut b)?
        };

        if b.is_empty() {
            return Ok(Lock::new(lock_type, primary, ts, ttl, None));
        }

        let flag = b.read_u8()?;
        assert_eq!(
            flag, SHORT_VALUE_PREFIX,
            "invalid flag [{:?}] in write",
            flag
        );

        let len = b.read_u8()?;
        if len as usize != b.len() {
            panic!(
                "short value len [{}] not equal to content len [{}]",
                len,
                b.len()
            );
        }

        Ok(Lock::new(lock_type, primary, ts, ttl, Some(b.to_vec())))
    }
}
```

```rust
// storage/mvcc/reader/mod.rs
 /// The return type is `(locks, is_remain)`. `is_remain` indicates whether there MAY be
    /// remaining locks that can be scanned.
    pub fn scan_locks<F>(
        &mut self,
        start: Option<&Key>,
        filter: F,
        limit: usize,
    ) -> Result<(Vec<(Key, Lock)>, bool)>
    where
        F: Fn(&Lock) -> bool,
    {
        self.create_lock_cursor()?;
        let cursor = self.lock_cursor.as_mut().unwrap();
        let ok = match start {
            Some(ref x) => cursor.seek(x, &mut self.statistics.lock)?,
            None => cursor.seek_to_first(&mut self.statistics.lock),
        };
        if !ok {
            return Ok((vec![], false));
        }
        let mut locks = Vec::with_capacity(limit);
        while cursor.valid()? {
            let key = Key::from_encoded_slice(cursor.key(&mut self.statistics.lock));
            let lock = Lock::parse(cursor.value(&mut self.statistics.lock))?;
            if filter(&lock) {
                locks.push((key, lock));
                if limit > 0 && locks.len() == limit {
                    return Ok((locks, true));
                }
            }
            cursor.next(&mut self.statistics.lock);
        }
        self.statistics.lock.processed += locks.len();
        // If we reach here, `cursor.valid()` is `false`, so there MUST be no more locks.
        Ok((locks, false))
    }

```

```rust
// storage/txn/process.rs

fn process_read_impl<E: Engine>(
    sched_ctx: &mut SchedContext<E>,
    mut cmd: Command,
    snapshot: E::Snap,
    statistics: &mut Statistics,
) -> Result<ProcessResult> {
    let tag = cmd.tag();
    match cmd {
        Command::MvccByKey { ref ctx, ref key } => {
            let mut reader = MvccReader::new(
                snapshot,
                Some(ScanMode::Forward),
                !ctx.get_not_fill_cache(),
                None,
                None,
                ctx.get_isolation_level(),
            );
            let result = find_mvcc_infos_by_key(&mut reader, key, u64::MAX);
            statistics.add(reader.get_statistics());
            let (lock, writes, values) = result?;
            Ok(ProcessResult::MvccKey {
                mvcc: MvccInfo {
                    lock,
                    writes,
                    values,
                },
            })
        }
        Command::MvccByStartTs { ref ctx, start_ts } => {
            let mut reader = MvccReader::new(
                snapshot,
                Some(ScanMode::Forward),
                !ctx.get_not_fill_cache(),
                None,
                None,
                ctx.get_isolation_level(),
            );
            match reader.seek_ts(start_ts)? {
                Some(key) => {
                    let result = find_mvcc_infos_by_key(&mut reader, &key, u64::MAX);
                    statistics.add(reader.get_statistics());
                    let (lock, writes, values) = result?;
                    Ok(ProcessResult::MvccStartTs {
                        mvcc: Some((
                            key,
                            MvccInfo {
                                lock,
                                writes,
                                values,
                            },
                        )),
                    })
                }
                None => Ok(ProcessResult::MvccStartTs { mvcc: None }),
            }
        }
        // Scans locks with timestamp <= `max_ts`
        Command::ScanLock {
            ref ctx,
            max_ts,
            ref start_key,
            limit,
            ..
        } => {
            let mut reader = MvccReader::new(
                snapshot,
                Some(ScanMode::Forward),
                !ctx.get_not_fill_cache(),
                None,
                None,
                ctx.get_isolation_level(),
            );
            let result = reader.scan_locks(start_key.as_ref(), |lock| lock.ts <= max_ts, limit);
            statistics.add(reader.get_statistics());
            let (kv_pairs, _) = result?;
            let mut locks = Vec::with_capacity(kv_pairs.len());
            for (key, lock) in kv_pairs {
                let mut lock_info = LockInfo::new();
                lock_info.set_primary_lock(lock.primary);
                lock_info.set_lock_version(lock.ts);
                lock_info.set_key(key.into_raw()?);
                locks.push(lock_info);
            }
            sched_ctx
                .command_keyread_duration
                .with_label_values(&[tag])
                .observe(locks.len() as f64);
            Ok(ProcessResult::Locks { locks })
        }
        Command::ResolveLock {
            ref ctx,
            ref mut txn_status,
            ref scan_key,
            ..
        } => {
            let mut reader = MvccReader::new(
                snapshot,
                Some(ScanMode::Forward),
                !ctx.get_not_fill_cache(),
                None,
                None,
                ctx.get_isolation_level(),
            );
            let result = reader.scan_locks(
                scan_key.as_ref(),
                |lock| txn_status.contains_key(&lock.ts),
                RESOLVE_LOCK_BATCH_SIZE,
            );
            statistics.add(reader.get_statistics());
            let (kv_pairs, has_remain) = result?;
            sched_ctx
                .command_keyread_duration
                .with_label_values(&[tag])
                .observe(kv_pairs.len() as f64);
            if kv_pairs.is_empty() {
                Ok(ProcessResult::Res)
            } else {
                let next_scan_key = if has_remain {
                    // There might be more locks.
                    kv_pairs.last().map(|(k, _lock)| k.clone())
                } else {
                    // All locks are scanned
                    None
                };
                Ok(ProcessResult::NextCommand {
                    cmd: Command::ResolveLock {
                        ctx: ctx.clone(),
                        txn_status: mem::replace(txn_status, Default::default()),
                        scan_key: next_scan_key,
                        key_locks: kv_pairs,
                    },
                })
            }
        }
        Command::Pause { duration, .. } => {
            thread::sleep(Duration::from_millis(duration));
            Ok(ProcessResult::Res)
        }
        _ => panic!("unsupported read command"),
    }
}
```
