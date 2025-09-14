// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Collections.Generic;
using Ycs.Contracts;

namespace Ycs.Core
{
    public class YDeepEventArgs
    {
        public YDeepEventArgs(IList<IYEvent> events, ITransaction transaction)
        {
            Events = events;
            Transaction = transaction;
        }

        public IList<IYEvent> Events { get; }
        public ITransaction Transaction { get; }
    }
}
