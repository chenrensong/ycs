// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Diagnostics;
using System.IO;
using Ycs.Lib0;

namespace Ycs.Contracts
{
    public struct StructID : IEquatable<StructID>
    {
        /// <summary>
        /// Client id.
        /// </summary>
        public long Client;

        /// <summary>
        /// Unique per client id, continuous number.
        /// </summary>
        public long Clock;

        public StructID(long client, long clock)
        {
            Debug.Assert(client >= 0, "Client should not be negative, as it causes client encoder to fail");
            Debug.Assert(clock >= 0);

            Client = client;
            Clock = clock;
        }

        public bool Equals(StructID other)
        {
            return Client == other.Client && Clock == other.Clock;
        }

        public static bool Equals(StructID? a, StructID? b)
        {
            return a == null && b == null || a != null && b != null && a.Value.Equals(b.Value);
        }

        public void Write(Stream writer)
        {
            writer.WriteVarUint((uint)Client);
            writer.WriteVarUint((uint)Clock);
        }

        public static StructID Read(Stream reader)
        {
            var client = reader.ReadVarUint();
            var clock = reader.ReadVarUint();
            Debug.Assert(client >= 0 && clock >= 0);
            return new StructID(client, clock);
        }
    }
}
