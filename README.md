# PS Projektna Naloga

Izidor Golčar, Branko Raičković

### Build

`./scripts/build.sh`\
\
Skript _build.sh_ prvo namesti potrebne knjižnice in
z ukazom `go generate` generira potrebne datoteke
(implementacija protobuf in grpc specifikacije, stringer).\
\
Ko so pripravljene vse izhhodiščne datoteke,
skript generira 4 binarne datoteke:

- `control_service`
- `data_service`
- `client_cli`
- `client_tui`

### Struktura

Projekt je organiziran v mapi [cmd](cmd) in [internal](internal).
Cmd vsebuje vhodne točke različnih komponent sistema.
internal se deli na module:

1. [client](internal/client)\
   Odjemalec s **TUI** vmesnikom za interakcijo s klepetalnico

2. [control](internal/control)\
   Implementacija <u>nadzorne ravnine</u> klepetalnice s podporo za **RAFT** protokol

3. [data](internal/data)\
   Implementacija <u>podatkovne ravnine</u> klepetalnice z **bločno replikacijo** sporočil med vozlišči

### Pregled

★ Zanesljiva bločna replikacija
: Modul [chain](internal/data/chain) poskrbi za varno replikacijo sporočil.

★ RAFT replikacija nadzorne ravnine
: Nadzorna ravnina se replicira z uporabo knjižnice [Hashicorp Raft](https://github.com/hashicorp/raft) 
ali pa z lastno implementacijo RAFT protokola.

★ Abstrakcija med podatkovno in nadzorno ravnino
: Podatkovna ravnina je odgovorna izključno za podatke, nadzorna pa za nadzor stanja podatkovne verige.\
Podatki nikoli, niti med prevezovanjem, ne prehajajo med ravninama. To omogoča protokol rokovanja med podatkovni vozlišči ([chain/handshake](internal/data/chain/handshake)).

★ TUI in CLI odjemalca
: Na voljo sta odjemalca za interakcijo z ukazno vrstico ali grafičnim vmesnikom v terminalu.

★ Pregleden ukazni vmesnik
: Ukazni vmesnik je organiziran s paketom [Cobra](https://github.com/spf13/cobra)

★ Pokritost z unit testi
: `go test ./...`

## Demonstracija

1. Za vzpostavitev sistema moramo zgolj zagnati nadzorno ravnino.
   ```shell
   DATA_NODE_EXEC = <path/to/data_service/executable>
   
   # Launch bootstrapped control node
   control_service launch \
      --node-id node1 \
      --raft-addr 127.0.0.1:5301 \
      --http-addr :8301 \
      --rpc-addr :8080 \
      --data-exec $DATA_NODE_EXEC \
      --bootstrap
   
   ... launch additional nodes without bootstrap flag
   ```
   
   Zagnana vozlišča povežemo v gručo:
   ```shell
   control_service link \
      --src "localhost:8301" \
      --target "127.0.0.1:5302" \
      --target-id node2
   ```
   
   Za nadzor stanja sistema lahko uporabimo ukaz `state`:
   ```shell
   control_service state \```
         --addr "localhost:8301"
   ```

2. Ko nadzorna ravnina vzpostavi podatkovno verigo lahko začnemo uporabljati odjemalce
    - **CLI**
    ```shell
   client_cli -addr <naslov poljubnega nadzornega vozlišča>
   ```
    - **TUI**
    ```shell
   client_tui -addr <naslov poljubnega nadzornega vozlišča>
   ```