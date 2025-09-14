// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

namespace Ycs.Contracts
{
    public struct DeleteItem
    {
        public readonly long Clock;
        public readonly long Length;

        public DeleteItem(long clock, long length)
        {
            Clock = clock;
            Length = length;
        }
    }
}
