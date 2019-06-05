# Firestore

+ acid transactions
+ multi-region replication

## Costs
100 000 writes TODO

> Cloud Firestore is a realtime database, meaning that clients can listen to data in Cloud Firestore and be notified in real time as it changes. This feature lets you build responsive apps that work regardless of network latency or Internet connectivity.

https://cloud.google.com/firestore/docs/overview


> You may notice that documents look a lot like JSON. In fact, they basically are. There are some differences (for example, documents support extra data types and are limited in size to 1 MB), but in general, you can treat documents as lightweight JSON records.

> Now that you have chat rooms, decide how to store your messages. You might not want to store them in the chat room's document. Documents in Cloud Firestore should be lightweight, and a chat room could contain a large number of messages. However, you can create additional collections within your chat room's document, as subcollections.

https://cloud.google.com/firestore/docs/data-model

no clob/blob?
https://cloud.google.com/firestore/docs/concepts/data-types


