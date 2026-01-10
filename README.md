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

★ Polna podpora za RAFT protokol
: Modul [raft](internal/control/raft) implementira vse potrebne funkcionalnosti za RAFT protokol.

★ Abstrakcija med podatkovno in nadzorno ravnino
: Podatkovna ravnina je odgovorna izključno za podatke, nadzorna pa za nadzor stanja podatkovne verige.\
Podatki nikoli, niti med prevezovanjem, ne prehajajo med ravninama.

★ TUI in CLI odjemalca
: Na voljo sta odjemalca za interakcijo z ukazno vrstico ali grafičnim vmesnikom v terminalu.

★ Pokritost z unit testi
: `go test ./...`

## Demonstracija

1. Za vzpostavitev sistema moramo zgolj zagnati nadzorno ravnino.
   ```shell
   ./build/control_service -config ... todo
   ```

2. Ko nadzorna ravnina vzpostavi podatkovno verigo lahko začnemo uporabljati odjemalce
    - **CLI**
    ```shell
   ./build/client_cli -addr <naslov nadzornega vozlišča>
   ```
    - **TUI**
    ```shell
   ./build/client_tui -addr <naslov nadzornega vozlišča>
   ```