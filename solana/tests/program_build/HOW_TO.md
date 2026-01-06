# How to copy build files

1. Clone `https://github.com/Ethernal-Tech/skyline-solana-programs` into separate folder.
2. If you have original key in `target/deploy` then `anchor build` would be enough, if not then:
   - run `solana keygen new -o target/deploy/skyline_program-keypair.json`
   - then run `anchor keys sync`
   - copy from `src/skyline-program/lib.rs` from `declare_id!(ID)` this id and paste it in `solana/skyline_program` in this project
3. Run `anchor build` in `skyline-solana-programs`.
4. Copy everything from `target/deploy` into `solana/test/program_build`.
