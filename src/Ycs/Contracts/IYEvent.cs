// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Collections.Generic;

namespace Ycs.Contracts
{
    public interface IYEvent
    {
        ChangesCollection Changes { get; }
        IAbstractType CurrentTarget { get; set; }
        IReadOnlyCollection<object> Path { get; }
        IAbstractType Target { get; set; }
        ITransaction Transaction { get; set; }
    }
}