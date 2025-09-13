// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.IO;

namespace Ycs.Contracts
{
    public interface IDSDecoder : IDisposable
    {
        Stream Reader { get; }

        void ResetDsCurVal();
        long ReadDsClock();
        long ReadDsLength();
    }

    public interface IUpdateDecoder : IDSDecoder
    {
        StructID ReadLeftId();
        StructID ReadRightId();
        long ReadClient();
        byte ReadInfo();
        string ReadString();
        bool ReadParentInfo();
        uint ReadTypeRef();
        int ReadLength();
        object ReadAny();
        byte[] ReadBuffer();
        string ReadKey();
        object ReadJson();
    }
}
