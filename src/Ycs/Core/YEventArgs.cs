// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using Ycs.Contracts;

namespace Ycs.Core
{
    public class YEventArgs
    {
        public YEventArgs(IYEvent evt, ITransaction transaction)
        {
            Event = evt;
            Transaction = transaction;
        }

        public IYEvent Event { get; }
        public ITransaction Transaction { get; }
    }
}
