.. ChatterBox documentation master file, created by
   sphinx-quickstart on Wed Apr 19 14:08:36 2017.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

Single chat
===========

.. toctree::
   :maxdepth: 2
   :caption: Contents:

Send Text Message
-----------------
send message from user vahid by id ``1`` to user saeed by id ``2``

topic for vahid: ``users/1``

topic for saeed: ``users/2``

vahid client send a text message to saeed topic ``users/2``:

.. code-block:: javascript

   {
      type: 1, // type of packet, 1 is message
      message: {
         id: 1223123, // randome incremental integer 64 bit
         type: 100, // type of message, 100 is text only
         chat: {
            code: "single-1:2", // single-MINid:MAXid
            type: 1,
         },
         from: {
            id: 1, // user id
            first_name: "vahid",
            last_name: "chakoshy", // optional
            username: "vahid" // optional
         },
         date: 1492597799, // current unix time stamp
         text: "salam",
      }
   }

Send Typing Message
-------------------

vahid client send a typing message to saeed topic ``users/2``:

.. code-block:: javascript

   {
      type: 2, // type of packet, 2 is typing
      message: {
         chat: {
            code: "single-1:2", // single-MINid:MAXid
            type: 1,
         },
         from: {
            id: 1, // user id
            first_name: "vahid",
            last_name: "chakoshy", // optional
            username: "vahid" // optional
         },
      }
   }

Send Seen Message
-------------------

vahid client send a seend message to saeed topic ``users/2``:

.. code-block:: javascript

   {
      type: 3, // type of packet, 3 is seen message
      message: {
         id: 1223124, // seen message id
         chat: {
            code: "single-1:2", // single-MINid:MAXid
            type: 1,
         },
         from: {
            id: 1, // user id
            first_name: "vahid",
            last_name: "chakoshy", // optional
            username: "vahid" // optional
         },
      }
   }

