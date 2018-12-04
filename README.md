# dwdtrayicon

https://sebastiankind.de/

Dieses Programm ermöglicht es, einige Wetterdaten:={Radarbilder, Satellitenbilder} aus pcmet bequem aus seiner Taskleiste/Panel/Trayiconhost zu betrachten, ohne zuvor erst auf pcment -> https://www.flugwetter.de/ zu wechseln. Um das Programm benutzen zu können, sind selbstverständlich gültige Zugangsdaten erforderlich.

![example.png](https://raw.githubusercontent.com/zeppel13/dwdtrayicon/master/example.png)

Zum Kompilieren muss go installiert sein. Außerdem besteht eine
Abhängigkeit zur Bibliothek systray und gogtk3 .

```
#systray installieren
go get github.com/getlantern/systray

```


```
# in den Projektornder wechseln und
go build
# zum Kompilieren ausführen
# Im Anschluss kann _radar_ gestartet werden
radar -user USERNAME -passwd PASSWORD --viewer IMAGEVIEWER
```
